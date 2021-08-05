package device

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	deviceCommand := &cobra.Command{
		Use:   "device",
		Short: "Device commands.",
		Long:  "Device commands.",
	}

	deviceCommand.AddCommand(initCreateCommand())
	deviceCommand.AddCommand(initListCommand())

	return deviceCommand
}
