package main

import (
	"github.com/habiliai/agentruntime/network"

	"github.com/habiliai/agentruntime/errors"
	"github.com/spf13/cobra"
)

func newNetworkThreadCmd() *cobra.Command {
	threadArgs := &struct {
		rpcEndpoint string
	}{}
	cmd := &cobra.Command{
		Use:     "thread",
		Short:   "Thread commands",
		Aliases: []string{"threads"},
	}

	cmd.PersistentFlags().StringVar(&threadArgs.rpcEndpoint, "rpc-endpoint", "http://127.0.0.1:9080/rpc",
		"Specify the address of the network server",
	)

	createCmd := func() *cobra.Command {
		kvargs := &struct {
			instruction string
		}{}
		cmd := &cobra.Command{
			Use:   "create",
			Short: "Create a thread",
			RunE: func(cmd *cobra.Command, args []string) error {
				networkClient := network.NewJsonRpcClient(threadArgs.rpcEndpoint)

				resp, err := networkClient.CreateThread(cmd.Context(), &network.CreateThreadRequest{
					Instruction: kvargs.instruction,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to create thread")
				}

				println("Thread created with ID:", resp.ThreadId)

				return nil
			},
		}

		cmd.Flags().StringVar(&kvargs.instruction, "instruction", "", "Instruction for the thread")

		return cmd
	}

	cmd.AddCommand(
		createCmd(),
	)

	return cmd
}
