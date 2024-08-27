package account

import (
	"regexp"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a collaborator to your account",
	Long: `Invite a collaborator to your account. This allows them to assume
access into your account and perform actions on your behalf.

Example Usages:
  chariot account add --email research@praetorian.com
`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		emailRegexp := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegexp.MatchString(email) {
			cmd.PrintErrf("Failed to add account: invalid email address, %s\n", email)
			return
		}

		account := model.NewAccount(Client.Username, email, "", nil)
		err := Client.Accounts.Add(account)
		if err != nil {
			cmd.PrintErrf("Failed to add account: %v\n", err)
			return
		}
		cmd.Printf("Linked %s to %s successfully\n", email, Client.Username)
	},
}

func init() {
	addCmd.Flags().String("email", "", "Email of the account (required)")
	addCmd.MarkFlagRequired("email")
}
