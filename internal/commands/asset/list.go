package asset

import (
	"encoding/json"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all assets",
	Long: `Return the list of all assets in the Chariot database.

Example Usages:
	chariot asset list
	chariot asset list --details
	chariot asset list --filter "key=example.com"`,

	Run: func(cmd *cobra.Command, args []string) {
		showDetails, _ := cmd.Flags().GetBool("details")
		filter, _ := cmd.Flags().GetString("filter")

		assets, err := Client.Assets.List()
		if err != nil {
			cmd.Printf("Failed to list assets: %v\n", err)
			return
		}

		if filter != "" {
			assets = model.FilterAssetsByKey(assets, filter)
		}

		for _, asset := range assets {
			if showDetails {
				json, _ := json.Marshal(asset)
				cmd.Printf("%s\n", json)
			} else {
				cmd.Printf("%s\n", asset.Key)
			}
		}
	},
}

func init() {
	listCmd.Flags().Bool("details", false, "Show detailed information about each asset")
	listCmd.Flags().String("filter", "", "Filter the assets list (e.g., --filter 'key=example.com')")
}
