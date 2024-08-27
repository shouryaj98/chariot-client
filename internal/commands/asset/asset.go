package asset

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var assetCmd = &cobra.Command{
	Use:   "asset",
	Short: "Manage assets within your attack surface",
	Long: `Assets are the core building blocks of your attack surface.
Manage your assets by adding, updating, deleting, and listing them.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	assetCmd.AddCommand(
		addCmd,
		updateCmd,
		deleteCmd,
		listCmd,
	)
}

var Client *sdk.Chariot

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return assetCmd
}
