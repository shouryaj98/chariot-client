package risk

import (
	"github.com/spf13/cobra"

	"github.com/praetorian-inc/chariot-client/pkg/sdk"
)

// riskCmd represents the risk command
var riskCmd = &cobra.Command{
	Use:   "risk",
	Short: "Manage your risks in Chariot",
	Long: `Valid Severity Levels:
	"I" - Informational
	"L" - Low
	"M" - Medium
	"H" - High
	"C" - Critical`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

var Client *sdk.Chariot

func init() {
	riskCmd.AddCommand(
		addCmd,
		updateCmd,
		deleteCmd,
		listCmd,
		proofCmd,
		definitionCmd,
	)
}

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return riskCmd
}
