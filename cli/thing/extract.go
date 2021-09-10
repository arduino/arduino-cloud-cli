package thing

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var extractFlags struct {
	id      string
	outfile string
	format  string
}

func initExtractCommand() *cobra.Command {
	extractCommand := &cobra.Command{
		Use:   "extract",
		Short: "Extract a template from a thing",
		Long:  "Extract a template from a Arduino IoT Cloud thing and save it in a file",
		RunE:  runExtractCommand,
	}
	extractCommand.Flags().StringVarP(&extractFlags.id, "id", "i", "", "Thing ID")
	extractCommand.Flags().StringVarP(&extractFlags.outfile, "outfile", "o", "", "Template file destination path")
	extractCommand.Flags().StringVar(
		&extractFlags.format,
		"format",
		"yaml",
		"Format of template file, can be {json|yaml}. Default is 'yaml'",
	)

	extractCommand.MarkFlagRequired("id")
	return extractCommand
}

func runExtractCommand(cmd *cobra.Command, args []string) error {
	logrus.Infof("Extracting template from thing %s\n", extractFlags.id)

	params := &thing.ExtractParams{
		ID:     extractFlags.id,
		Format: extractFlags.format,
	}
	if extractFlags.outfile != "" {
		params.Outfile = &extractFlags.outfile
	}

	err := thing.Extract(params)
	if err != nil {
		feedback.Errorf("Error during template extraction: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Template successfully extracted")
	return nil
}
