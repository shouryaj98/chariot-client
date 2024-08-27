package job

import (
	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a job",
	Long: `Add a job to the Chariot database. This command requires a capability for the job and a key for the asset to queue the job for.

Example Usages:
  chariot job add --capability nuclei --key #asset#example.com#1.2.3.4
`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("capability")
		key, _ := cmd.Flags().GetString("key")

		asset, err := model.GetAssetFromKey(key)
		if err != nil {
			cmd.PrintErrf("Failed to get asset from key: %v\n", err)
			return
		}

		job := model.NewJob(source, *asset)
		err = Client.Jobs.Add(job)
		if err != nil {
			cmd.PrintErrf("Failed to add job: %v\n", err)
			return
		}

		cmd.Printf("Job added successfully\n")
	},
}

func init() {
	addCmd.Flags().String("capability", "", "Capability to run")
	addCmd.Flags().String("key", "", "Key of the asset to queue a job for")
	addCmd.MarkFlagRequired("key")
	addCmd.MarkFlagRequired("source")
}
