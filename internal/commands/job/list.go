package job

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jobs",
	Long: `Return the list of all jobs in the Chariot database.

Example Usages:
	chariot job list`,
	Run: func(cmd *cobra.Command, args []string) {
		capability, _ := cmd.Flags().GetString("capability")
		status, _ := cmd.Flags().GetString("status")
		details, _ := cmd.Flags().GetBool("details")

		jobs, err := Client.Jobs.List()
		if err != nil {
			cmd.Printf("Failed to list jobs: %v\n", err)
			return
		}

		for _, job := range jobs {
			if capability != "" && strings.ToLower(job.Source) != strings.ToLower(capability) {
				continue
			}

			if status != "" && string(job.Status[1]) != status {
				continue
			}

			if details {
				cmd.Printf("%s\n", job.Raw())
			} else {
				cmd.Printf("%s,%s,%s\n", job.DNS, job.Source, job.Status)
			}
		}
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		if status != "" {
			switch status {
			case "Q", "R", "F", "P":
				return nil
			default:
				return fmt.Errorf("invalid job status level: %s", status)
			}
		}
		return nil
	},
}

func init() {
	listCmd.Flags().String("capability", "", "Filter the list of risks by capability")
	listCmd.Flags().String("status", "", "Filter the list of risks by status (Q, R, F, P)")
	listCmd.Flags().Bool("details", false, "Show detailed information about each risk")
}
