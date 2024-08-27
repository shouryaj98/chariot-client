package asset

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an asset",
	Long: `Specify the key of the asset to delete with --key. This will flag the asset as
deleted in the Chariot database and prevent further scanning.

Example Usages:
  chariot asset delete --key #asset#example.com#1.2.3.4`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")

		asset, err := model.GetAssetFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get asset from key: %v\n", err)
			return
		}

		err = Client.Assets.Delete(*asset)
		if err != nil {
			cmd.PrintErrf("Failed to delete asset: %v\n", err)
			return
		}
		cmd.Printf("Asset %s deleted successfully\n", key)
	},
}

func init() {
	deleteCmd.Flags().String("key", "", "Key of the asset to remove")
	deleteCmd.MarkFlagRequired("key")
}
