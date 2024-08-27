package attribute

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an attribute",
	Long: `Specify the key of the attribute to delete with --key.

Example Usages:
  chariot attribute delete --key #attribute#technology#HTML Forms#asset#example.com#12.34.56.78`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")

		attribute, err := model.GetAttributeFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get attribute from key: %v\n", err)
			return
		}

		err = Client.Attributes.Delete(*attribute)
		if err != nil {
			cmd.PrintErrf("Failed to delete attribute: %v\n", err)
			return
		}
		cmd.Printf("Attribute %s deleted successfully\n", key)
	},
}

func init() {
	deleteCmd.Flags().String("key", "", "Key of the attribute to remove")
	deleteCmd.MarkFlagRequired("key")
}
