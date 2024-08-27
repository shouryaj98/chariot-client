package account

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a collaborator from your account",
	Long: `Remove a collaborator from your account. This will revoke their access to your account.

Example Usages:
  chariot account delete --email research@praetorian.com`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")

		account := model.NewAccount(Client.Username, email, "", nil)

		err := Client.Accounts.Delete(account)
		if err != nil {
			cmd.PrintErrf("Failed to unlink colaborator: %v\n", err)
			return
		}
		cmd.Printf("Unlinked %s from %s successfully\n", email, Client.Username)
	},
}

func init() {
	deleteCmd.Flags().String("email", "", "Email of the account (required)")
	deleteCmd.MarkFlagRequired("email")
}
