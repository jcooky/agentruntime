package main

import (
	"context"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"github.com/habiliai/agentruntime/internal/msgutils"
	"github.com/habiliai/agentruntime/internal/stringslices"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/thread"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"io"
	"slices"
	"strconv"
)

func newConnectCmd() *cobra.Command {
	flags := &struct {
		addr     string
		noSecure bool
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

			conn, err := grpcutils.NewClient(flags.addr, !flags.noSecure)
			if err != nil {
				return errors.WithStack(err)
			}
			defer conn.Close()

			threadManager := thread.NewThreadManagerClient(conn)
			agentNetwork := network.NewAgentNetworkClient(conn)

			thr, err := threadManager.GetThread(ctx, &thread.GetThreadRequest{
				ThreadId: uint32(threadId),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to get thread with id %d", threadId)
			}

			pterm.DefaultLogger.Info("thr: ", []pterm.LoggerArgument{{Key: "thr", Value: thr}})

			secondary := pterm.ThemeDefault.SecondaryStyle

			interrupted := false
			textInput := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithOnInterruptFunc(func() {
				interrupted = true
			})

			var lastMessageId uint32
			for {
				userInput, err := textInput.Show("> You")
				if err != nil {
					return err
				}

				if interrupted {
					break
				}
				if msg, err := threadManager.AddMessage(ctx, &thread.AddMessageRequest{
					ThreadId: uint32(threadId),
					Content:  userInput,
					Sender:   "USER",
				}); err != nil {
					return errors.Wrap(err, "failed to add message")
				} else {
					lastMessageId = msg.MessageId
				}

				agentMentions := msgutils.ExtractMentions(userInput)
				runtimeInfo, err := agentNetwork.GetAgentRuntimeInfo(ctx, &network.GetAgentRuntimeInfoRequest{
					Names: agentMentions,
				})
				if err != nil {
					return errors.Wrap(err, "failed to get agent runtime info")
				}

				if len(runtimeInfo.AgentRuntimeInfo) == 0 {
					secondary.Println("< Agent: ", "No agent found")
				} else {
					var (
						eg errgroup.Group
					)
					for _, info := range runtimeInfo.AgentRuntimeInfo {
						names := stringslices.IntersectIgnoreCase(info.AgentNames, agentMentions)
						if len(names) == 0 {
							continue
						}

						eg.Go(func() error {
							conn, err := grpcutils.NewClient(info.Addr, info.Secure)
							if err != nil {
								return errors.Wrapf(err, "failed to create gRPC client. addr: %s", info.Addr)
							}
							defer conn.Close()

							runtimeClient := runtime.NewAgentRuntimeClient(conn)
							if _, err := runtimeClient.Run(ctx, &runtime.RunRequest{
								ThreadId:   uint32(threadId),
								AgentNames: names,
							}); err != nil {
								return errors.Wrapf(err, "failed to run agent. addr: %s", info.Addr)
							}

							return nil
						})
					}
					if err := eg.Wait(); err != nil {
						return err
					}
				}

				{
					var messages []*thread.Message
					ctx, cancel := context.WithCancel(ctx)
					defer cancel()

					if stream, err := threadManager.GetMessages(ctx, &thread.GetMessagesRequest{
						ThreadId: uint32(threadId),
						Order:    thread.GetMessagesRequest_LATEST,
					}); err != nil {
						return errors.Wrap(err, "failed to get messages")
					} else {
						for interrupt := false; !interrupt; {
							msg, err := stream.Recv()
							if err == io.EOF {
								break
							} else if err != nil {
								return errors.Wrapf(err, "failed to receive message")
							}

							for _, m := range msg.Messages {
								if m.Id == lastMessageId {
									interrupt = true
									break
								}
								messages = append(messages, m)
							}
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
	f.BoolVarP(&flags.noSecure, "no-secure", "s", false, "Specify connect without SSL/TLS")
	f.StringVarP(&flags.addr, "addr", "A", "127.0.0.1:9080", "Specify the address of the server")

	return cmd
}
