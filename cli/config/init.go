package config

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initFlags struct {
	destDir   string
	overwrite bool
	format    string
}

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize a configuration file with default values",
		Long:  "Initialize an Arduino IoT Cloud CLI configuration file with default values",
		Run:   runInitCommand,
	}

	initCommand.Flags().StringVar(&initFlags.destDir, "dest-dir", ".", "Sets where to save the configuration file.")
	initCommand.Flags().BoolVar(&initFlags.overwrite, "overwrite", false, "Overwrite existing config file.")
	initCommand.Flags().StringVar(&initFlags.format, "config-format", "yaml", "Format of the configuration file, can be {yaml|json}")

	return initCommand
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Initializing a config file in folder: %s", initFlags.destDir)

	params := &config.InitParams{
		DestDir:   initFlags.destDir,
		Overwrite: initFlags.overwrite,
		Format:    initFlags.format,
	}

	err := config.Init(params)
	if err != nil {
		feedback.Errorf("Error during config init: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Config file successfully initialized")
}
