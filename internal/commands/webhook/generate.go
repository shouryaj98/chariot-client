package webhook

import (
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new webhook. This will replace older webooks, if they exist.",
	Long: `Create a webhook for posting content into Chariot.

Example Usages:
  chariot webhook generate
`,
	Run: func(cmd *cobra.Command, args []string) {
		webhook, err := Client.Accounts.AddWebhook()
		if err != nil {
			cmd.PrintErrf("Failed to generate webhook: %v\n", err)
			return
		}
		cmd.Printf("Webhook generated: %s\n", webhook)
	},
}
