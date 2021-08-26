package thing

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/spf13/cobra"
)

var deleteFlags struct {
	id string
}

func initDeleteCommand() *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete a thing",
		Long:  "Delete a thing from Arduino IoT Cloud",
		RunE:  runDeleteCommand,
	}
	deleteCommand.Flags().StringVarP(&deleteFlags.id, "id", "i", "", "Thing ID")
	deleteCommand.MarkFlagRequired("id")
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Deleting thing %s\n", deleteFlags.id)

	params := &thing.DeleteParams{ID: deleteFlags.id}
	err := thing.Delete(params)
	if err != nil {
		return err
	}

	fmt.Println("Thing successfully deleted")
	return nil
}
