package config

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands.",
		Long:  "Configuration commands.",
	}

	configCommand.AddCommand(initInitCommand())

	return configCommand
}
