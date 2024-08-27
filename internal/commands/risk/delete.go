package risk

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a risk",
	Long:  `Specify the key of the risk to delete with --key. This will flag the risk as deleted in the Chariot database.`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")

		risk, err := model.GetRiskFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get risk from key: %v\n", err)
			return
		}

		err = Client.Risks.Delete(*risk)
		if err != nil {
			cmd.PrintErrf("Failed to delete risk: %v\n", err)
			return
		}

		cmd.Printf("Risk %s deleted successfully\n", key)
	},
}

func init() {
	deleteCmd.Flags().String("key", "", "Key of the risk to remove")
	deleteCmd.MarkFlagRequired("key")
}
