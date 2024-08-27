package file

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all files",
	Long: `Return the list of all files in the Chariot database.

Example Usages:
	chariot file list
	chariot file list --filter "proofs/example.com"`,
	Run: func(cmd *cobra.Command, args []string) {
		filter, _ := cmd.Flags().GetString("filter")
		proofs, _ := cmd.Flags().GetBool("proofs")
		definitions, _ := cmd.Flags().GetBool("definitions")

		files, err := Client.Files.List()
		if err != nil {
			cmd.PrintErrf("Failed to list files: %v\n", err)
			return
		}

		filtered := make([]model.File, 0)
		if proofs {
			filtered = append(filtered, model.FilterFilesByName(files, "proofs/")...)
		}

		if definitions {
			filtered = append(filtered, model.FilterFilesByName(files, "definitions/")...)
		}

		if filter != "" {
			filtered = model.FilterFilesByName(filtered, filter)
		}

		if !proofs && !definitions && filter == "" {
			filtered = files
		}

		for _, file := range filtered {
			cmd.Printf("%s\n", file.Name)
		}
	},
}

func init() {
	listCmd.Flags().String("filter", "", "Filter the files by name (e.g., --filter 'example.com')")
	listCmd.Flags().Bool("proofs", false, "Show only proofs")
	listCmd.Flags().Bool("definitions", false, "Show only definitions")
}
