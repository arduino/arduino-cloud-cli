package ota

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/ota"
	"github.com/spf13/cobra"
)

var uploadFlags struct {
	deviceID string
	file     string
}

func initUploadCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: "OTA upload",
		Long:  "OTA upload on a device of Arduino IoT Cloud",
		RunE:  runUploadCommand,
	}

	uploadCommand.Flags().StringVarP(&uploadFlags.deviceID, "device-id", "d", "", "Device ID")
	uploadCommand.Flags().StringVarP(&uploadFlags.file, "file", "", "", "OTA file")

	uploadCommand.MarkFlagRequired("device-id")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runUploadCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Uploading binary %s to device %s\n", uploadFlags.file, uploadFlags.deviceID)

	params := &ota.UploadParams{
		DeviceID: uploadFlags.deviceID,
		File:     uploadFlags.file,
	}
	err := ota.Upload(params)
	if err != nil {
		return err
	}

	fmt.Println("Upload successfully started")
	return nil
}
