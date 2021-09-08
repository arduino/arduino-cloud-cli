package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/cli/config"
	"github.com/arduino/iot-cloud-cli/cli/device"
	"github.com/arduino/iot-cloud-cli/cli/thing"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cliFlags struct {
	verbose      bool
	outputFormat string
}

func Execute() {
	cli := &cobra.Command{
		Use:              "arduino-cloud-cli",
		Short:            "Arduino Cloud CLI.",
		Long:             "Arduino Cloud Command Line Interface (arduino-cloud-cli).",
		PersistentPreRun: preRun,
	}

	cli.AddCommand(config.NewCommand())
	cli.AddCommand(device.NewCommand())
	cli.AddCommand(thing.NewCommand())

	cli.PersistentFlags().BoolVarP(&cliFlags.verbose, "verbose", "v", false, "Print the logs on the standard output.")
	cli.PersistentFlags().StringVar(&cliFlags.outputFormat, "format", "text", "The output format, can be {text|json}.")

	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func parseFormatString(arg string) (feedback.OutputFormat, bool) {
	f, found := map[string]feedback.OutputFormat{
		"json": feedback.JSON,
		"text": feedback.Text,
	}[arg]

	return f, found
}

func preRun(cmd *cobra.Command, args []string) {
	// decide whether we should log to stdout
	if cliFlags.verbose {
		// if we print on stdout, do it in full colors
		logrus.SetOutput(colorable.NewColorableStdout())
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	} else {
		logrus.SetOutput(ioutil.Discard)
	}

	// normalize the format strings
	cliFlags.outputFormat = strings.ToLower(cliFlags.outputFormat)
	// check the right output format was passed
	format, found := parseFormatString(cliFlags.outputFormat)
	if !found {
		feedback.Error("Invalid output format: " + cliFlags.outputFormat)
		os.Exit(errorcodes.ErrBadCall)
	}
	// use the output format to configure the Feedback
	feedback.SetFormat(format)

	if cliFlags.outputFormat != "text" {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			feedback.Error("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(errorcodes.ErrBadCall)
		})
	}
}
