package config

import (
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/iot-cloud-cli/command/config"
	"github.com/sirupsen/logrus"
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
		feedback.Error("Error during config: provide either a yaml file or credentials")
		os.Exit(errorcodes.ErrGeneric)
	}

	conf := viper.New()

	if configFlags.file != "" {
		file := paths.New(configFlags.file)
		filename := strings.TrimSuffix(file.String(), file.Ext())
		conf.SetConfigName(filename)
		conf.SetConfigType(strings.Trim(file.Ext(), "."))
		conf.AddConfigPath(".")
		err := conf.ReadInConfig()
		if err != nil {
			feedback.Errorf("Error during config: fatal error config file: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}

	} else {
		conf.BindPFlag("client", cmd.Flag("client"))
		conf.BindPFlag("secret", cmd.Flag("secret"))
	}

	err := config.Config(conf)
	if err != nil {
		feedback.Errorf("Error during config: storing config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Configuration file updated")
	return nil
}
