package ota

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	otaCommand := &cobra.Command{
		Use:   "ota",
		Short: "Over The Air.",
		Long:  "Over The Air firmware update.",
	}

	otaCommand.AddCommand(initUploadCommand())

	return otaCommand
}
