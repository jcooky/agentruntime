package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/msgutils"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/thread"
	"github.com/mokiat/gog"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newConnectCmd() *cobra.Command {
	flags := &struct {
		url string
	}{}
	cmd := &cobra.Command{
		Use: "connect <thread-id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) != 1 {
				return errors.New("thread-id is required")
			}

			threadId, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.Wrapf(err, "failed to convert thread-id %s to int", args[0])
			}

			threadClient := thread.NewJsonRpcClient(flags.url)
			networkClient := network.NewJsonRpcClient(flags.url)

			thr, err := threadClient.GetThread(ctx, &thread.GetThreadRequest{
				ThreadId: uint32(threadId),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to get thread with id %d", threadId)
			}

			logger := pterm.DefaultLogger
			logger.Info("thr: ", []pterm.LoggerArgument{{Key: "thr", Value: thr}})

			secondary := pterm.ThemeDefault.SecondaryStyle

			interrupted := false
			textInput := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithOnInterruptFunc(func() {
				interrupted = true
			})

			var (
				lastMessageId uint32
			)
			for {
				userInput, err := textInput.Show("> You")
				if err != nil {
					return err
				}

				if interrupted {
					break
				}

				if reply, err := threadClient.AddMessage(ctx, &thread.AddMessageRequest{
					ThreadId: uint32(threadId),
					Content:  userInput,
					Sender:   "USER",
				}); err != nil {
					return errors.Wrapf(err, "failed to add message")
				} else {
					lastMessageId = reply.MessageId
				}

				agentMentions := msgutils.ExtractMentions(userInput)
				reply, err := networkClient.GetAgentRuntimeInfo(ctx, &network.GetAgentRuntimeInfoRequest{
					Names: agentMentions,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to get agent runtime info")
				}

				runtimeInfoAgg := make(map[string][]*network.AgentRuntimeInfo)
				for _, info := range reply.AgentRuntimeInfo {
					runtimeInfoAgg[info.Addr] = append(runtimeInfoAgg[info.Addr], info)
				}

				if len(reply.AgentRuntimeInfo) == 0 {
					secondary.Println("< Agent: ", "No agent found")
				} else {
					for addr, info := range runtimeInfoAgg {
						names := gog.Map(info, func(i *network.AgentRuntimeInfo) string {
							return i.Info.Name
						})
						runtimeClient := runtime.NewJsonRpcClient(addr)

						if _, err := runtimeClient.Run(ctx, &runtime.RunRequest{
							ThreadId:   uint32(threadId),
							AgentNames: names,
						}); err != nil {
							logger.Error(fmt.Sprintf("failed to run agent. err: %v, agentNames: '%v'", err, names))
						}
					}
				}

				{
					var (
						messages []*thread.Message
						cursor   uint32
					)
					ctx, cancel := context.WithCancel(ctx)
					defer cancel()

					for interrupt := false; !interrupt; {
						reply, err := threadClient.GetMessages(ctx, &thread.GetMessagesRequest{
							ThreadId: uint32(threadId),
							Order:    "latest",
							Cursor:   cursor,
						})
						if err != nil {
							return errors.Wrapf(err, "failed to get messages")
						}

						for _, m := range reply.Messages {
							if m.Id == lastMessageId {
								interrupt = true
								break
							}
							messages = append(messages, m)
						}
					}

					if len(messages) == 0 {
						continue
					}
					slices.Reverse(messages)

					for _, m := range messages {
						if m.Sender == "USER" {
							return errors.New("user message not expected")
						} else {
							secondary.Printf("< Agent(@%s): %s\n", m.Sender, m.Content)
						}
					}
				}
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&flags.url, "url", "A", "http://127.0.0.1:9080", "Specify the address of the server")

	return cmd
}
