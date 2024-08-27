package commands

import (
	"fmt"
	"github.com/praetorian-inc/chariot-client/internal/commands/account"
	"github.com/spf13/viper"
	"os"

	"github.com/praetorian-inc/chariot-client/internal/commands/asset"
	"github.com/praetorian-inc/chariot-client/internal/commands/attribute"
	"github.com/praetorian-inc/chariot-client/internal/commands/file"
	"github.com/praetorian-inc/chariot-client/internal/commands/job"
	"github.com/praetorian-inc/chariot-client/internal/commands/risk"
	"github.com/praetorian-inc/chariot-client/internal/commands/search"
	"github.com/praetorian-inc/chariot-client/internal/commands/webhook"
	"github.com/praetorian-inc/chariot-client/pkg/sdk"

	"github.com/spf13/cobra"
)

var (
	accountOverride string
	profileOverride string
	cfgFile         string
	Chariot         *sdk.Chariot
)

var rootCmd = &cobra.Command{
	Use:   "chariot",
	Short: "Command line interface for interacting with Chariot",
	Long:  ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chariot.yaml)")
	rootCmd.PersistentFlags().Lookup("config").Hidden = true

	rootCmd.PersistentFlags().String("profile", "", "profile name from your keychain (stored at $HOME/.praetorian/keychain.ini) to use")
	rootCmd.PersistentFlags().String("account", "", "account to perform actions for")

	initConfig()

	rootCmd.AddCommand(
		account.Cmd(Chariot),
		asset.Cmd(Chariot),
		attribute.Cmd(Chariot),
		file.Cmd(Chariot),
		job.Cmd(Chariot),
		risk.Cmd(Chariot),
		search.Cmd(Chariot),
		webhook.Cmd(Chariot),
	)
}

func initConfig() {
	rootCmd.ParseFlags(os.Args)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".chariot")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	profileOverride, _ = rootCmd.PersistentFlags().GetString("profile")
	accountOverride, _ = rootCmd.PersistentFlags().GetString("account")
	Chariot = sdk.NewClient(profileOverride)
	profileName := viper.Get("profile")
	if profileName != nil && profileOverride != "" {
		Chariot = sdk.NewClient(profileName.(string))
	}
	if accountOverride != "" {
		Chariot.SetAccount(accountOverride)
	}
}
