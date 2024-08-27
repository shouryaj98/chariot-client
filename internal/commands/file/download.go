package file

import (
	"os"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download the file from the specified file path",
	Long: `Download the file from the specified file path,

Example Usages:
	chariot file download --name "path/to/fileNameOnChariot.txt" --path /path/to/downloadedFile.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		path, _ := cmd.Flags().GetString("path")
		b, err := Client.DownloadFile(model.NewFile(name))
		if err != nil {
			cmd.Printf("Failed to read file: %v\n", err)
			return
		}
		if path == "" {
			path = name
		}
		os.WriteFile(path, b, 0644)
		cmd.Printf("Saved file at %s\n", path)
	},
}

func init() {
	downloadCmd.Flags().String("name", "", "Name of the file to read")
	downloadCmd.Flags().String("path", "", "Path to save the file to")
	downloadCmd.MarkFlagRequired("name")
}
