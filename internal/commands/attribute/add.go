package attribute

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an attribute",
	Long: `Add an attribute (key-value pair) to an asset or risk. 

Example Usages:
	chariot attribute add --name "OS" --value "Windows 10" --key "#asset#example.com#1.2.3.4"`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")
		key, _ := cmd.Flags().GetString("key")

		attribute := model.NewAttribute(name, value, key)
		err := Client.Attributes.Add(attribute)
		if err != nil {
			cmd.PrintErrf("Failed to add attribute: %v\n", err)
			return
		}

		cmd.Printf("Attribute added successfully\n")
	},
}

func init() {
	addCmd.Flags().String("name", "", "Name of attribute")
	addCmd.Flags().String("value", "", "Value of attribute")
	addCmd.Flags().String("key", "", "The key of the target risk or asset")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("value")
	addCmd.MarkFlagRequired("key")
}
