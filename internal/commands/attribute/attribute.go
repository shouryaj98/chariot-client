package attribute

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var attributeCmd = &cobra.Command{
	Use:   "attribute",
	Short: "Manage attributes tied to assets and risks",
	Long:  `Chariot collects metadata about assets, known as attributes. These key/value pairs describe specific properties`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

var Client *sdk.Chariot

func init() {
	attributeCmd.AddCommand(
		addCmd,
		deleteCmd,
		listCmd,
	)
}

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return attributeCmd
}
