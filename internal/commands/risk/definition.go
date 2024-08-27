package risk

import (
	"os"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var definitionCmd = &cobra.Command{
	Use:   "definition",
	Short: "Manage the definitions for a risk",
	Long: `Definitions are markdown files which are displayed by the Chariot UI to provide context for a risk. Use this command to upload and download definitions for specific risks.

Example Usages:
  chariot risk definition upload --name "sql-injection" --file sql_injection_md_file
  chariot risk definition download --name "sql-injection"`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Args: cobra.MinimumNArgs(1),
}

var defUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a definition for a risk",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		file, _ := cmd.Flags().GetString("file")

		data, err := os.ReadFile(file)
		if err != nil {
			cmd.PrintErrf("Failed to read file: %v\n", err)
			return
		}

		err = Client.UploadDefinition(model.Risk{Name: name}, data)
		if err != nil {
			cmd.PrintErrf("Failed to upload definition: %v\n", err)
			return
		}

		cmd.Printf("Definition uploaded successfully\n")
	},
}

var defDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a definition for a risk",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		b, err := Client.DownloadDefinition(model.Risk{Name: name})
		if err != nil {
			cmd.PrintErrf("Failed to download definition: %v\n", err)
			return
		}

		cmd.Print(string(b))
	},
}

func init() {
	definitionCmd.AddCommand(
		defUploadCmd,
		defDownloadCmd,
	)

	defUploadCmd.Flags().String("name", "", "Name of the risk")
	defUploadCmd.Flags().String("file", "", "File to upload")
	defUploadCmd.MarkFlagRequired("name")
	defUploadCmd.MarkFlagRequired("file")

	defDownloadCmd.Flags().String("name", "", "Name of the risk")
	defDownloadCmd.MarkFlagRequired("name")
}
