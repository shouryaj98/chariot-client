package webhook

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Generate or display your Chariot webhook",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

var Client *sdk.Chariot

func init() {
	jobCmd.AddCommand(
		generateCmd,
		showCmd,
	)
}

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return jobCmd
}
