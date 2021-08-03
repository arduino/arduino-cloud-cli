package cli

import (
	"fmt"
	"os"

	"github.com/arduino/iot-cloud-cli/cli/config"
	"github.com/arduino/iot-cloud-cli/cli/device"
	"github.com/arduino/iot-cloud-cli/cli/thing"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(config.NewCommand())
	rootCmd.AddCommand(device.NewCommand())
	rootCmd.AddCommand(thing.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
