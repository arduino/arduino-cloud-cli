package device

import (
	"fmt"

	"github.com/bcmi-labs/iot-cloud-cli/command/config"
	"github.com/spf13/cobra"
)

var createFlags struct {
	port string
	name string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long:  "Create a device for Arduino IoT Cloud",
		RunE:  runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Device name")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Creating device with name %s\n", createFlags.name)

	conf, _ := config.Retrieve()
	fmt.Println(conf)

	return nil
}
