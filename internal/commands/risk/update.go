package risk

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the status or severity of a risk",
	Long:  `When a risk needs to have its status changed (such as closing a risk or modifying its severity), this is the command to use.`,
	Run: func(cmd *cobra.Command, args []string) {
		key := cmd.Flag("key").Value.String()
		status := cmd.Flag("status").Value.String()

		risk, err := model.GetRiskFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get risk from key: %v\n", err)
			return
		}

		// we should probably validate the status but man that'd be a
		// huge case tree...
		risk.Status = status

		err = Client.Risks.Update(*risk)
		if err != nil {
			cmd.PrintErrf("Failed to update risk: %v\n", err)
			return
		}
		cmd.Printf("Risk %s updated successfully\n", key)
	},
}

func init() {
	updateCmd.Flags().String("status", "", "Status of the risk, can be...")
	updateCmd.Flags().String("key", "", "Key of the asset")
	updateCmd.MarkFlagRequired("status")
	updateCmd.MarkFlagRequired("key")
}
