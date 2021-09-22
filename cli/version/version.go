package version

import (
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	v "github.com/arduino/arduino-cloud-cli/version"
	"github.com/spf13/cobra"
)

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Shows version number of Arduino Cloud CLI.",
		Long:    "Shows the version number of Arduino Cloud CLI which is installed on your system.",
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run:     run,
	}
}

func run(cmd *cobra.Command, args []string) {
	feedback.Print(v.VersionInfo)
}
