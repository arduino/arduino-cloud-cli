package config

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/cli/feedback"
	paths "github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFlags struct {
	file   string
	client string
	secret string
}

func NewCommand() *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "Set the configuration file",
		Long:  "Set the configuration file to access Arduino IoT Cloud",
		RunE:  runConfigCommand,
	}
	configCommand.Flags().StringVarP(&configFlags.file, "file", "f", "", "Existing configuration yaml file")
	configCommand.Flags().StringVarP(&configFlags.client, "client", "c", "", "Client ID")
	configCommand.Flags().StringVarP(&configFlags.secret, "secret", "s", "", "Secret ID")
	return configCommand
}

func runConfigCommand(cmd *cobra.Command, args []string) error {
	if configFlags.file == "" && (configFlags.client == "" || configFlags.secret == "") {
		return fmt.Errorf("%s", "Provide either a yaml file or credentials\n")
	}

	if configFlags.file != "" {
		file := paths.New(configFlags.file)
		filename := strings.TrimSuffix(file.String(), file.Ext())
		conf := viper.New()
		conf.SetConfigName(filename)
		conf.SetConfigType(strings.Trim(file.Ext(), "."))
		conf.AddConfigPath(".")
		err := conf.ReadInConfig()
		if err != nil {
			feedback.Errorf("Fatal error config file:  %v", err)
			return err
		}

		if err := conf.WriteConfigAs("config.yaml"); err != nil {
			feedback.Errorf("Cannot create config file: %v", err)
			return err
		}

	} else {
		conf := viper.New()
		conf.BindPFlag("client", cmd.Flag("client"))
		conf.BindPFlag("secret", cmd.Flag("secret"))

		if err := conf.WriteConfigAs("config.yaml"); err != nil {
			feedback.Errorf("Cannot create config file: %v", err)
			return err
		}
	}

	fmt.Println("Configuration file updated")
	return nil
}