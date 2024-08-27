package asset

import (
	"strings"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an asset's priority",
	Long: `Update an asset's priority to comprehensive, standard, discover, or frozen.
For example:
  chariot update asset --priority comprehensive --key #asset#example.com#1.2.3.4`,

	Run: func(cmd *cobra.Command, args []string) {
		key := cmd.Flag("key").Value.String()
		priority := cmd.Flag("priority").Value.String()

		asset, err := model.GetAssetFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get asset from key: %v\n", err)
			return
		}

		switch strings.ToLower(priority) {
		case "comprehensive":
			asset.Status = model.ActiveHigh
		case "standard":
			asset.Status = model.Active
		case "discover":
			asset.Status = model.ActiveLow
		case "frozen":
			asset.Status = model.Frozen
		default:
			cmd.Help()
			return
		}

		err = Client.Assets.Update(*asset)
		if err != nil {
			cmd.PrintErrf("Failed to update asset: %v\n", err)
			return
		}
		cmd.Printf("Asset %s updated successfully\n", key)
	},
}

func init() {
	updateCmd.Flags().String("priority", "", "Priority of the asset - can be comprehensive, standard, discover, or frozen (required)")
	updateCmd.Flags().String("key", "", "Key of the asset (required)")
	updateCmd.MarkFlagRequired("priority")
	updateCmd.MarkFlagRequired("key")
}
