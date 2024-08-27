package search

import (
	"encoding/json"
	"reflect"

	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var Client *sdk.Chariot

func Cmd(client *sdk.Chariot) *cobra.Command {
	Client = client
	return searchCmd
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search chariot with a provided search term",
	Long: `Search Chariot using one of the following search mechanisms: Searching by key, dns, source, status, name, or ip

Example Usages:
	chariot search --term name:test
	chariot search --term source:provided --details
	chariot search --term dns:example.com`,

	Run: func(cmd *cobra.Command, args []string) {
		showDetails, _ := cmd.Flags().GetBool("details")

		results, err := Client.Search(cmd.Flag("term").Value.String())
		if err != nil {
			cmd.Printf("Failed to search: %v\n", err)
			return
		}

		var items []interface{}

		for _, a := range results.Assets {
			items = append(items, a)
		}
		for _, a := range results.Accounts {
			items = append(items, a)
		}
		for _, a := range results.Attributes {
			items = append(items, a)
		}
		for _, j := range results.Jobs {
			items = append(items, j)
		}
		for _, r := range results.Risks {
			items = append(items, r)
		}
		for _, f := range results.Files {
			items = append(items, f)
		}

		for _, item := range items {
			if showDetails {
				json, _ := json.Marshal(item)
				cmd.Printf("%s\n", json)
			} else {
				var key string
				v := reflect.ValueOf(item)
				field := v.FieldByName("Key")
				if field.IsValid() && field.Kind() == reflect.String {
					key = field.String()
				} else {
					continue
				}
				cmd.Printf("%s\n", key)
			}
		}
	},
}

func init() {
	searchCmd.Flags().Bool("details", false, "Show detailed information about each item")
	searchCmd.Flags().String("term", "", "Search term to use (required)")

	searchCmd.MarkFlagRequired("term")
}
