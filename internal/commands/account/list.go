package account

import (
	"encoding/json"
	"strings"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts",
	Long: `Retrieve and display a list of all accounts linked to the current profile in the Chariot database.

Example Usages:
	chariot account list --members
	chariot account list --linked --details
	chariot account list --members --filter 'praetorian.com'`,

	Run: func(cmd *cobra.Command, args []string) {
		showDetails, _ := cmd.Flags().GetBool("details")
		members, _ := cmd.Flags().GetBool("members")
		linked, _ := cmd.Flags().GetBool("linked")
		filter, _ := cmd.Flags().GetString("filter")

		accounts, err := Client.Accounts.List()
		if err != nil {
			cmd.Printf("Failed to list accounts: %v\n", err)
			return
		}

		if members {
			accounts = filterAccounts(accounts, Client.Username, false)
		} else if linked {
			accounts = filterAccounts(accounts, Client.Username, true)
		}

		if filter != "" {
			// TODO: Fix filtering to filter members correctly
			accounts = filterAccounts(accounts, filter, false)
		}

		for _, account := range accounts {
			if account.Member == "settings" {
				continue
			}
			if showDetails {
				json, _ := json.Marshal(account)
				cmd.Printf("%s\n", json)
			} else {
				if account.Name == Client.Username {
					cmd.Printf("%s\n", account.Member)
				} else {
					cmd.Printf("%s\n", account.Name)
				}
			}
		}
	},
}

func filterAccounts(accounts []model.Account, filter string, invert bool) []model.Account {
	filtered := make([]model.Account, 0)
	for _, account := range accounts {
		filterStr := account.Name
		if (strings.Contains(filterStr, filter) && !invert) || (!strings.Contains(filterStr, filter) && invert) {
			filtered = append(filtered, account)
		}
	}
	return filtered
}

func init() {
	listCmd.Flags().SortFlags = false
	listCmd.Flags().BoolP("members", "m", false, "List the members of your account")
	listCmd.Flags().BoolP("linked", "l", false, "List the accounts you are a member of")
	listCmd.MarkFlagsMutuallyExclusive("members", "linked")

	listCmd.Flags().BoolP("details", "d", false, "Show detailed information about each asset")
	listCmd.Flags().String("filter", "", "Filter the assets list (e.g., --filter 'praetorian.com')")
}
