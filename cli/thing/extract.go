package thing

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/spf13/cobra"
)

var extractFlags struct {
	id      string
	outfile string
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
	extractCommand.MarkFlagRequired("id")
	return extractCommand
}

func runExtractCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Extracting template from thing %s\n", extractFlags.id)

	params := &thing.ExtractParams{ID: extractFlags.id}
	if extractFlags.outfile != "" {
		params.Outfile = &extractFlags.outfile
	}

	err := thing.Extract(params)
	if err != nil {
		return err
	}

	fmt.Println("Template successfully extracted")
	return nil
}
