package file

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Handling files stored in Chariot",
	Long:  `Chariot stores files for risk definitions, proof of exploitation, and other supporting information. Use this command to upload, download, and list these files.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	fileCmd.AddCommand(
		uploadCmd,
		downloadCmd,
		listCmd,
	)
}

var Client *sdk.Chariot

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return fileCmd
}
