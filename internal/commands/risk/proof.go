package risk

import (
	"os"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var proofCmd = &cobra.Command{
	Use:   "proof",
	Short: "Manage proof tied to a risk exposure",
	Long: `Proof of exploitation is generated by the platform when capabilities identify risks. This command can be used to manually upload proof for risks that may have been added manually or to download proof for existing risks in the Chariot platform.

Example Usages:
  chariot risk proof upload --key "#risk#target.site#sql-injection" --file proof_file
  chariot risk proof download --key "#risk#target.site#sql-injection"`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

var proofUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a proof for a risk",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		file, _ := cmd.Flags().GetString("file")

		risk, err := model.GetRiskFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get risk from key: %v\n", err)
			return
		}

		data, err := os.ReadFile(file)
		if err != nil {
			cmd.PrintErrf("Failed to read file: %v\n", err)
			return
		}

		err = Client.UploadPoE(*risk, data)
		if err != nil {
			cmd.PrintErrf("Failed to upload PoE: %v\n", err)
			return
		}

		cmd.Printf("Proof uploaded successfully\n")
	},
}

var proofDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a proof for a risk",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")

		risk, err := model.GetRiskFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get risk from key: %v\n", err)
			return
		}

		b, err := Client.DownloadPoE(*risk)
		if err != nil {
			cmd.PrintErrf("Failed to download PoE: %v\n", err)
			return
		}

		cmd.Print(string(b))
	},
}

func init() {
	proofCmd.AddCommand(
		proofUploadCmd,
		proofDownloadCmd,
	)

	proofUploadCmd.Flags().String("key", "", "Key of the risk")
	proofUploadCmd.Flags().String("file", "", "File to upload")
	proofUploadCmd.MarkFlagRequired("key")
	proofUploadCmd.MarkFlagRequired("file")

	proofDownloadCmd.Flags().String("key", "", "Key of the risk")
	proofDownloadCmd.MarkFlagRequired("key")
}
