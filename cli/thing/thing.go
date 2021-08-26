package thing

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	thingCommand := &cobra.Command{
		Use:   "thing",
		Short: "Thing commands.",
		Long:  "Thing commands.",
	}

	thingCommand.AddCommand(initCreateCommand())

	return thingCommand
}
