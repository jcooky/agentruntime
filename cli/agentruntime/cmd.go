package agentruntime

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentruntime",
		Short: "Agent runtime by HabiliAI",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	cmd.AddCommand(
		newAgentCmd(),
		newThreadCmd(),
		newServeCmd(),
	)

	return cmd
}
