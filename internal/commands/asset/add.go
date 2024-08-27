package asset

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an asset",
	Long: `Add an asset to the Chariot database. This command requires a DNS name for the asset.
Optionally, a name can be provided to give the asset more specific information, such as IP address.
If no name is provided, the DNS name will be used as the name.

Example asset:

- acme.com: Domain name
- 8.8.8.8: IP Addresses
- 8.8.8.0/24: CIDR Ranges
- https://github.com/acme-corp: GitHub Org

Example Usages:
  chariot asset add --dns example.com
  chariot asset add --dns example.com --name 1.2.3.4
`,
	Run: func(cmd *cobra.Command, args []string) {
		dns, _ := cmd.Flags().GetString("dns")
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = dns
		}

		asset := model.NewAsset(dns, name)
		err := Client.Assets.Add(asset)
		if err != nil {
			cmd.PrintErrf("Failed to add asset: %v\n", err)
			return
		}
		cmd.Printf("Asset %s added successfully\n", asset.Key)
	},
}

func init() {
	addCmd.Flags().String("dns", "", "DNS of the asset")
	addCmd.Flags().String("name", "", "Name of the asset")
	addCmd.MarkFlagRequired("dns")
}
