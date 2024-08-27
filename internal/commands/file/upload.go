package file

import (
	"os"

	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to the server",
	Long: `Upload a file to the server at the specified path via name.

Example Usages:
	chariot file upload --name "fileNameOnChariot.txt" --file /path/to/file.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		file, _ := cmd.Flags().GetString("file")

		data, err := os.ReadFile(file)
		if err != nil {
			cmd.PrintErrf("Failed to read file: %v\n", err)
			return
		}

		err = Client.Upload(name, data)
		if err != nil {
			cmd.PrintErrf("Failed to upload file: %v\n", err)
			return
		}

		cmd.Printf("File %s uploaded successfully\n", name)
	},
}

func init() {
	uploadCmd.Flags().String("name", "", "Name to store the file as in Chariot")
	uploadCmd.Flags().String("file", "", "Path to the file to upload")
	uploadCmd.MarkFlagRequired("name")
	uploadCmd.MarkFlagRequired("file")
}
