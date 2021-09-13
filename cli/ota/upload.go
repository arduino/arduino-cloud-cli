package ota

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/ota"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var uploadFlags struct {
	deviceID string
	file     string
	deferred bool
}

func initUploadCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: "OTA upload",
		Long:  "OTA upload on a device of Arduino IoT Cloud",
		Run:   runUploadCommand,
	}

	uploadCommand.Flags().StringVarP(&uploadFlags.deviceID, "device-id", "d", "", "Device ID")
	uploadCommand.Flags().StringVarP(&uploadFlags.file, "file", "", "", "Binary file (.bin) to be uploaded")
	uploadCommand.Flags().BoolVar(&uploadFlags.deferred, "deferred", false, "Perform a deferred OTA. It can take up to 1 week.")

	uploadCommand.MarkFlagRequired("device-id")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runUploadCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Uploading binary %s to device %s", uploadFlags.file, uploadFlags.deviceID)

	params := &ota.UploadParams{
		DeviceID: uploadFlags.deviceID,
		File:     uploadFlags.file,
	}
	err := ota.Upload(params)
	if err != nil {
		feedback.Errorf("Error during ota upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Upload successfully started")
}
