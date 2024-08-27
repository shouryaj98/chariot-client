package risk

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new risk to an asset",
	Long: `Add a new risk to the Chariot database for a given asset. This command requires a DNS name for the asset.
Manually entered risks will be tracked but not automatically monitored.

Example Usages:
  chariot risk add --dns example.com --risk "SQL Injection"
`,
	Run: func(cmd *cobra.Command, args []string) {
		dns, _ := cmd.Flags().GetString("dns")
		name, _ := cmd.Flags().GetString("name")

		risk := model.NewRisk(model.Asset{DNS: dns}, name)
		err := Client.Risks.Add(risk)
		if err != nil {
			cmd.PrintErrf("Failed to add risk: %v\n", err)
			return
		}

		cmd.Printf("Risk %s added successfully\n", risk.Key)
	},
}

func init() {
	addCmd.Flags().String("dns", "", "DNS of the asset")
	addCmd.Flags().String("name", "", "Name of the risk")
	addCmd.MarkFlagRequired("dns")
	addCmd.MarkFlagRequired("name")
}
