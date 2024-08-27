package job

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "View or create jobs in Chariot",
	Long:  `Jobs in Chariot track execution of capabilities performed by the Chariot platform. Use this command to view which capabilities have been launched recently or manually trigger a capability.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

var Client *sdk.Chariot

func init() {
	jobCmd.AddCommand(
		addCmd,
		listCmd,
	)
}

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return jobCmd
}
