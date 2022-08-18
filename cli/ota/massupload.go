// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package ota

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type massUploadFlags struct {
	deviceIDs []string
	tags      map[string]string
	file      string
	deferred  bool
	fqbn      string
}

func initMassUploadCommand() *cobra.Command {
	flags := &massUploadFlags{}
	massUploadCommand := &cobra.Command{
		Use:   "mass-upload",
		Short: "Mass OTA upload",
		Long:  "Mass OTA upload on devices of Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runMassUploadCommand(flags); err != nil {
				feedback.Errorf("Error during ota mass-upload: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	massUploadCommand.Flags().StringSliceVarP(&flags.deviceIDs, "device-ids", "d", nil,
		"Comma-separated list of device IDs to update")
	massUploadCommand.Flags().StringToStringVar(&flags.tags, "device-tags", nil,
		"Comma-separated list of tags with format <key>=<value>.\n"+
			"Perform an OTA upload on all devices that match the provided tags.\n"+
			"Mutually exclusive with '--device-ids'.",
	)
	massUploadCommand.Flags().StringVarP(&flags.file, "file", "", "", "Binary file (.bin) to be uploaded")
	massUploadCommand.Flags().BoolVar(&flags.deferred, "deferred", false, "Perform a deferred OTA. It can take up to 1 week.")
	massUploadCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "FQBN of the devices to update")
	massUploadCommand.MarkFlagRequired("file")
	massUploadCommand.MarkFlagRequired("fqbn")
	return massUploadCommand
}

func runMassUploadCommand(flags *massUploadFlags) error {
	logrus.Infof("Uploading binary %s", flags.file)

	params := &ota.MassUploadParams{
		DeviceIDs: flags.deviceIDs,
		Tags:      flags.tags,
		File:      flags.file,
		Deferred:  flags.deferred,
		FQBN:      flags.fqbn,
	}

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	resp, err := ota.MassUpload(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	// Put successful devices ahead
	sort.SliceStable(resp, func(i, j int) bool {
		return resp[i].Err == nil
	})

	feedback.PrintResult(massUploadResult{resp})

	var failed []string
	for _, r := range resp {
		if r.Err != nil {
			failed = append(failed, r.ID)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	failDevs := strings.Join(failed, ",")
	feedback.Printf(
		"You can try to perform the OTA again on the failed devices using the following command:\n"+
			"$ arduino-cloud-cli ota mass-upload --file %s --fqbn %s -d %s",
		params.File, params.FQBN, failDevs,
	)
	return nil
}

type massUploadResult struct {
	res []ota.Result
}

func (r massUploadResult) Data() interface{} {
	return r.res
}

func (r massUploadResult) String() string {
	if len(r.res) == 0 {
		return "No OTA done."
	}
	t := table.New()
	t.SetHeader("ID", "Result")
	for _, r := range r.res {
		outcome := "Success"
		if r.Err != nil {
			outcome = fmt.Sprintf("Fail: %s", r.Err.Error())
		}

		t.AddRow(
			r.ID,
			outcome,
		)
	}
	return t.Render()
}
