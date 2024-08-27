package attribute

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all attributes",
	Long: `Return the list of all attributes in the Chariot database.

Example Usages:
	chariot attribute list
	chariot attribute list --details`,
	Run: func(cmd *cobra.Command, args []string) {
		showDetails, _ := cmd.Flags().GetBool("details")

		attributes, err := Client.Attributes.List()
		if err != nil {
			cmd.Printf("Failed to list attributes: %v\n", err)
			return
		}

		for _, attribute := range attributes {
			if showDetails {
				json, _ := json.Marshal(attribute)
				cmd.Printf("%s\n", json)
			} else {
				cmd.Printf("%s\n", attribute.Key)
			}
		}
	},
}

func init() {
	listCmd.Flags().Bool("details", false, "Show detailed information about each asset")
	// listCmd.Flags().String("filter", "", "Filter the assets list (e.g., --filter 'key=example.com')")
}
