package cli

import (
	"fmt"
	"os"

	"github.com/bcmi-labs/iot-cloud-cli/cli/ping"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(ping.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
