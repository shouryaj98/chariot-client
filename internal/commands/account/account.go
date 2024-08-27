package account

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage accounts linked to your Chariot account",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
	},
	Args: cobra.MinimumNArgs(1),
}

var Client *sdk.Chariot

func init() {
	accountCmd.AddCommand(
		addCmd,
		deleteCmd,
		listCmd,
	)
}

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return accountCmd
}
