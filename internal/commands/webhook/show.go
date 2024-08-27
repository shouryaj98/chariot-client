package webhook

import (
	"encoding/base64"
	"github.com/spf13/cobra"
	"strings"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current webhook",
	Long: `Show the current webhook into chariot, if it has been created.

Example Usages:
  chariot webhook show
`,
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := Client.Accounts.List()

		for _, account := range accounts {
			if account.Member == "hook" {
				pin := account.Config["pin"]
				username := base64.StdEncoding.EncodeToString([]byte(Client.Username))
				encodedUsername := strings.TrimRight(username, "=")
				cmd.Printf("Webhook: %s/hook/%s/%s", Client.API, encodedUsername, pin)
				return
			}
		}
		if err != nil {
			cmd.PrintErrf("Failed to add webhook: %v\n", err)
			return
		}
		cmd.Printf("No existing webhook found, try using 'chariot webhook add'\n")
	},
}
