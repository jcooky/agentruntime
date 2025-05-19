package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/network"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type ChatUI struct {
	networkClient network.JsonRpcClient
	threadId      uint32
	lastMessageId uint32
	outputChan    chan *network.Message
	inputChan     chan string
	mu            sync.Mutex
	logger        pterm.Logger
	secondary     *pterm.Style
}

func newChatUI(networkClient network.JsonRpcClient, threadId uint32) *ChatUI {
	return &ChatUI{
		networkClient: networkClient,
		threadId:      threadId,
		outputChan:    make(chan *network.Message, 100),
		inputChan:     make(chan string, 10),
		logger:        pterm.DefaultLogger,
		secondary:     &pterm.ThemeDefault.SecondaryStyle,
	}
}

func (c *ChatUI) startMessagePolling(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pollMessages(ctx)
		}
	}
}

func (c *ChatUI) loadRecentMessages(ctx context.Context) {
	reply, err := c.networkClient.GetMessages(ctx, &network.GetMessagesRequest{
		ThreadId: c.threadId,
		Order:    "latest",
	})
	if err != nil {
		c.logger.Error("Failed to load recent messages", []pterm.LoggerArgument{{Key: "error", Value: err}})
		return
	}

	if len(reply.Messages) == 0 {
		fmt.Println("No previous messages found.")
		return
	}

	// Display recent messages (last 10 or all if less than 10)
	messages := reply.Messages
	if len(messages) > 10 {
		messages = messages[:10]
		fmt.Println("--- Showing last 10 messages ---")
	} else {
		fmt.Println("--- Previous messages ---")
	}

	for _, msg := range slices.Backward(messages) {
		var structuredMsg = struct {
			User string `json:"user"`
			Text string `json:"text"`
		}{}
		if err := json.Unmarshal([]byte(msg.Content), &structuredMsg); err != nil {
			structuredMsg.Text = msg.Content // fallback to plain text
		}

		if msg.Sender == "USER" {
			fmt.Printf("> You: %s\n", structuredMsg.Text)
		} else {
			fmt.Printf("< Agent(@%s): %s\n", msg.Sender, structuredMsg.Text)
		}

		// Update lastMessageId to the highest message ID seen
		if msg.Id > c.lastMessageId {
			c.lastMessageId = msg.Id
		}
	}

	fmt.Println("--- End of previous messages ---")
	fmt.Println()
}

func (c *ChatUI) pollMessages(ctx context.Context) {
	reply, err := c.networkClient.GetMessages(ctx, &network.GetMessagesRequest{
		ThreadId: c.threadId,
		Order:    "latest",
		Cursor:   0,
	})
	if err != nil {
		c.logger.Error("Failed to get messages", []pterm.LoggerArgument{{Key: "error", Value: err}})
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Find new messages (messages with ID greater than lastMessageId)
	for i := len(reply.Messages) - 1; i >= 0; i-- {
		msg := reply.Messages[i]
		if msg.Id > c.lastMessageId {
			select {
			case c.outputChan <- msg:
			case <-ctx.Done():
				return
			}
			c.lastMessageId = msg.Id
		}
	}
}

func (c *ChatUI) startInputHandler(ctx context.Context, cancel context.CancelFunc) {
	textInput := pterm.DefaultInteractiveTextInput.WithDefaultText("")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			userInput, err := textInput.Show("> You")
			if err != nil {
				c.logger.Error("Input error", []pterm.LoggerArgument{{Key: "error", Value: err}})
				continue
			}

			if userInput == "/quit" || userInput == "/exit" {
				cancel()
				return
			}

			select {
			case c.inputChan <- userInput:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *ChatUI) startOutputHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.outputChan:
			var structuredMsg = struct {
				User string `json:"user"`
				Text string `json:"text"`
			}{}
			if err := json.Unmarshal([]byte(msg.Content), &structuredMsg); err != nil {
				structuredMsg.Text = msg.Content // fallback to plain text
			}

			if msg.Sender == "USER" {
				fmt.Printf("> You: %s\n", structuredMsg.Text)
			} else {
				c.secondary.Printf("< Agent(@%s): %s\n", msg.Sender, structuredMsg.Text)
			}
		}
	}
}

func (c *ChatUI) handleUserInput(ctx context.Context, userInput string) {
	reply, err := c.networkClient.AddMessage(ctx, &network.AddMessageRequest{
		ThreadId: c.threadId,
		Content:  userInput,
		Sender:   "USER",
	})
	if err != nil {
		c.logger.Error("Failed to add message", []pterm.LoggerArgument{{Key: "error", Value: err}})
		return
	}

	c.mu.Lock()
	if reply.MessageId > c.lastMessageId {
		c.lastMessageId = reply.MessageId
	}
	c.mu.Unlock()
}

func (c *ChatUI) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Get initial thread info
	thr, err := c.networkClient.GetThread(ctx, &network.GetThreadRequest{
		ThreadId: c.threadId,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to get thread with id %d", c.threadId)
	}

	c.logger.Info("Connected to thread", []pterm.LoggerArgument{{Key: "thread", Value: thr}})
	fmt.Println("Chat started! Type '/quit' or '/exit' to exit.")
	fmt.Println("Loading recent messages...")
	fmt.Println()

	// Load recent messages first
	c.loadRecentMessages(ctx)

	// Start goroutines
	go c.startMessagePolling(ctx)
	go c.startOutputHandler(ctx)
	go c.startInputHandler(ctx, cancel)

	// Handle user inputs
	for {
		select {
		case <-ctx.Done():
			return nil
		case userInput := <-c.inputChan:
			c.handleUserInput(ctx, userInput)
		}
	}
}

func newConnectCmd() *cobra.Command {
	kvargs := &struct {
		rpcEndpoint string
	}{}
	cmd := &cobra.Command{
		Use: "connect <thread-id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("thread-id is required")
			}

			threadId, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.Wrapf(err, "failed to convert thread-id %s to int", args[0])
			}

			networkClient := network.NewJsonRpcClient(kvargs.rpcEndpoint)
			chatUI := newChatUI(networkClient, uint32(threadId))

			return chatUI.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&kvargs.rpcEndpoint, "rpc-endpoint", "http://127.0.0.1:9080/rpc",
		"Specify the address of the network server",
	)

	return cmd
}
