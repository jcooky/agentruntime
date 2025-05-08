package main

import (
	"fmt"
	"github.com/jcooky/go-din"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/thread"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newNetworkThreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "thread",
		Short:   "Thread commands",
		Aliases: []string{"threads"},
	}

	createCmd := func() *cobra.Command {
		kvargs := &struct {
			instruction string
		}{}
		cmd := &cobra.Command{
			Use:   "create",
			Short: "Create a thread",
			RunE: func(cmd *cobra.Command, args []string) error {
				c := din.NewContainer(cmd.Context(), din.EnvProd)

				threadManager, err := din.GetT[thread.Manager](c)
				if err != nil {
					return err
				}

				thread, err := threadManager.CreateThread(c, kvargs.instruction)
				if err != nil {
					return err
				}

				println("Thread created with ID:", thread.ID)

				return nil
			},
		}

		cmd.Flags().StringVar(&kvargs.instruction, "instruction", "", "Instruction for the thread")

		return cmd
	}

	addMessageCmd := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "add-message <thread-id> <message>",
			Short: "Add a message to a thread",
			RunE: func(cmd *cobra.Command, args []string) error {
				c := din.NewContainer(cmd.Context(), din.EnvProd)

				if len(args) < 2 {
					return errors.Errorf("thread-id and message are required")
				}

				threadId, err := strconv.Atoi(args[0])
				if err != nil {
					return errors.Errorf("thread-id must be an integer")
				}

				message := args[1]

				threadManager, err := din.GetT[thread.Manager](c)
				if err != nil {
					return err
				}

				if _, err := threadManager.AddMessage(c, uint(threadId), "USER", entity.MessageContent{
					Text: message,
				}); err != nil {
					return err
				}

				return nil
			},
		}

		return cmd
	}

	listCmd := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "list",
			Short: "List threads",
			RunE: func(cmd *cobra.Command, args []string) error {
				c := din.NewContainer(cmd.Context(), din.EnvProd)

				threadManager, err := din.GetT[thread.Manager](c)
				if err != nil {
					return err
				}

				var (
					cursor uint = 0
					limit  uint = 10
				)

				// Create a new screen
				screen, err := tcell.NewScreen()
				if err != nil {
					log.Fatalf("Error creating screen: %v", err)
				}
				// Initialize the screen
				if err := screen.Init(); err != nil {
					log.Fatalf("Error initializing screen: %v", err)
				}
				// Ensure the screen is finalized on exit
				defer screen.Fini()

				// Set the default style (reset background and foreground colors)
				defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
				screen.SetStyle(defStyle)
				screen.Clear()

				printText(screen, 0, 0, "Press Enter to load more posts. Press ESC to exit.")

				return listScreen(c, screen, ListScreenRequest{}, func() ([]string, error) {
					threads, err := threadManager.GetThreads(c, cursor, limit)
					if err != nil {
						return nil, err
					}
					if len(threads) == 0 {
						return nil, nil
					}

					messages := make([]string, 0, len(threads))
					for _, thd := range threads {
						messages = append(messages, fmt.Sprintf("Thread ID: %d", thd.ID))
						cursor = thd.ID
					}
					return messages, nil
				})
			},
		}

		return cmd
	}

	listMessagesCmd := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "list-messages <thread-id>",
			Short: "List messages in a thread",
			RunE: func(cmd *cobra.Command, args []string) error {
				c := din.NewContainer(cmd.Context(), din.EnvProd)

				if len(args) < 1 {
					return errors.Errorf("thread-id is required")
				}

				threadId, err := strconv.Atoi(args[0])
				if err != nil {
					return errors.Errorf("thread-id must be an integer")
				}

				threadManager, err := din.GetT[thread.Manager](c)
				if err != nil {
					return err
				}

				screen, err := tcell.NewScreen()
				if err != nil {
					log.Fatalf("Error creating screen: %v", err)
				}
				if err := screen.Init(); err != nil {
					log.Fatalf("Error initializing screen: %v", err)
				}
				defer screen.Fini()

				defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
				screen.SetStyle(defStyle)
				screen.Clear()

				var (
					cursor uint = 0
					limit  uint = 10
				)
				return listScreen(c, screen, ListScreenRequest{}, func() ([]string, error) {
					messages, err := threadManager.GetMessages(c, uint(threadId), "ASC", cursor, limit)
					if err != nil {
						return nil, err
					}

					if len(messages) == 0 {
						return nil, nil
					}

					res := make([]string, 0, len(messages))
					for _, msg := range messages {
						res = append(res, fmt.Sprintf("Message ID: %d, Text: %s, User: %s", msg.ID, msg.Content.Data().Text, msg.User))
						cursor = msg.ID
					}

					return res, nil
				})
			},
		}

		return cmd
	}

	cmd.AddCommand(
		createCmd(),
		addMessageCmd(),
		listCmd(),
		listMessagesCmd(),
	)

	return cmd
}
