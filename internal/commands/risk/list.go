package risk

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all risks",
	Long:  `Return the list of all risks in the Chariot database`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("source")
		severity, _ := cmd.Flags().GetString("severity")
		status, _ := cmd.Flags().GetString("status")
		details, _ := cmd.Flags().GetBool("details")

		risks, err := Client.Risks.List()
		if err != nil {
			cmd.Printf("Failed to list risks: %s\n", err)
			return
		}

		for _, risk := range risks {
			if severity != "" && risk.Severity() != severity {
				continue
			}
			if status != "" && string(risk.Status[0]) != status {
				continue
			}
			if source != "" && strings.ToLower(risk.Source) != strings.ToLower(source) {
				continue
			}
			if details {
				json, _ := json.Marshal(risk)
				cmd.Printf("%s\n", json)
			} else {
				cmd.Printf("%s\n", risk.Key)
			}
		}
	},

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		severity, _ := cmd.Flags().GetString("severity")
		if severity != "" {
			switch severity {
			case "I", "L", "M", "H", "C":
				return nil
			default:
				return fmt.Errorf("invalid severity level: %s", severity)
			}
		}
		status, _ := cmd.Flags().GetString("status")
		if status != "" {
			switch status {
			case "T", "O", "C", "M":
				return nil
			default:
				return fmt.Errorf("invalid status level: %s", status)
			}
		}
		return nil
	},
}

func init() {
	listCmd.Flags().String("source", "", "Filter the list of risks by source")
	listCmd.Flags().String("severity", "", "Filter the list of risks by severity (I, L, M, H, C)")
	listCmd.Flags().String("status", "", "Filter the list of risks by status (T, O, C, M)")
	listCmd.Flags().Bool("details", false, "Show detailed information about each risk")
}
