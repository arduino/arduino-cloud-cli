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

package provision

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	// old "github.com/arduino/arduino-cli/arduino/sketches"

	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/new/arduino/sketch"
	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	thingPropertiesFile = "thingProperties.h"
	tempPropertiesFile  = "thingProperties.temp"
)

var provisionFlags struct {
	name     string
	port     string
	fqbn     string
	template string
	sketch   string
}

type state struct {
	skip   bool
	device *device.DeviceInfo
	thing  *thing.ThingInfo
}

func (s *state) flush() {
	if s.skip {
		return
	}
	if s.device != nil {
		device.Delete(&device.DeleteParams{ID: &s.device.ID})
	}
	if s.thing != nil {
		thing.Delete(&thing.DeleteParams{ID: &s.thing.ID})
	}
}

func NewCommand() *cobra.Command {
	provisionCommand := &cobra.Command{
		Use:   "provision",
		Short: "Provision device, with thing and sketch",
		Long:  "Provision a device, attaching a thing to it and uploading a sketch to it.",
	}

	provisionCommand.AddCommand(initProvisionCommand())

	return provisionCommand
}

func initProvisionCommand() *cobra.Command {
	provisionCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long:  "Create a device for Arduino IoT Cloud",
		Run:   runProvisionCommand,
	}
	provisionCommand.Flags().StringVarP(&provisionFlags.port, "port", "p", "", "Device port")
	provisionCommand.Flags().StringVarP(&provisionFlags.name, "name", "n", "", "Device and thing name")
	provisionCommand.Flags().StringVarP(&provisionFlags.fqbn, "fqbn", "b", "", "Device fqbn")
	provisionCommand.Flags().StringVarP(&provisionFlags.sketch, "sketch", "s", "", "Path of the sketch to upload")
	provisionCommand.Flags().StringVarP(&provisionFlags.template, "template", "t", "", "Path of the thing template")

	provisionCommand.MarkFlagRequired("port")
	provisionCommand.MarkFlagRequired("fqbn")
	provisionCommand.MarkFlagRequired("name")
	provisionCommand.MarkFlagRequired("sketch")
	provisionCommand.MarkFlagRequired("template")
	return provisionCommand
}

func runProvisionCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Provisioning started")
	err := run()
	if err != nil {
		os.Exit(errorcodes.ErrGeneric)
	}
}

func run() error {
	s := state{skip: false}
	defer s.flush()

	deviceName := provisionFlags.name + "-device"
	thingName := provisionFlags.name + "-thing"

	logrus.Infof("Creating device with name %s", deviceName)
	dev, err := deviceCreate(deviceName)
	if err != nil {
		feedback.Errorf("Error during device create: %v", err)
		return err
	}
	s.device = dev

	logrus.Infof("Creating thing from template %s", provisionFlags.template)
	th, err := thing.Create(&thing.CreateParams{Name: &thingName, Template: provisionFlags.template})
	if err != nil {
		feedback.Errorf("Error during thing create: %v", err)
		return err
	}
	s.thing = th

	logrus.Info("Binding thing to device")
	err = thing.Bind(&thing.BindParams{ID: th.ID, DeviceID: dev.ID})
	if err != nil {
		feedback.Errorf("Error during thing bind: %v", err)
		return err
	}

	logrus.Info("Edit sketch thing-id")
	skPath := paths.New(provisionFlags.sketch)
	sk, err := sketch.New(skPath)
	if err != nil {
		feedback.Errorf("Error during sketch opening: %v", err)
		return err
	}
	err = editSketch(sk, th.ID)
	defer restoreSketch(sk)
	if err != nil {
		feedback.Errorf("Error during sketch edit: %v", err)
		return err
	}

	logrus.Info("Compile and upload the modified sketch")
	if err := uploadSketch(sk, provisionFlags.fqbn, "arduino:samd:mkrwifi1010"); err != nil {
		feedback.Errorf("Error during sketch upload: %v", err)
		return err
	}

	// Don't flush state if command was successful
	s.skip = true
	logrus.Infof("Device provisioned, device-id: %s, thing-id: %s", dev.ID, th.ID)
	return nil
}

func deviceCreate(name string) (*device.DeviceInfo, error) {
	params := &device.CreateParams{
		Name: name,
	}
	if provisionFlags.port != "" {
		params.Port = &provisionFlags.port
	}
	if provisionFlags.fqbn != "" {
		params.Fqbn = &provisionFlags.fqbn
	}
	return device.Create(params)
}

func editSketch(sk *sketch.Sketch, thingID string) (err error) {
	var propPath *paths.Path
	for _, f := range sk.AdditionalFiles {
		if f.Base() == thingPropertiesFile {
			propPath, err = f.Abs()
			if err != nil {
				return fmt.Errorf("retrieving '%s': %w", thingPropertiesFile, err)
			}
			break
		}
	}
	if propPath == nil {
		return fmt.Errorf("this is not a cloud sketch: '%s' not found", thingPropertiesFile)
	}

	// Change name of the original properties file. It will be restored later
	tempPath := sk.FullPath.Join(tempPropertiesFile)
	err = propPath.Rename(tempPath)
	if err != nil {
		return err
	}

	// Read the properties file content
	file, err := tempPath.Open()
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		return err
	}

	// Write the properties file with the new content
	thID := "setThingId(\"" + thingID + "\")"
	content = bytes.Replace(content, []byte("setThingId(THING_ID)"), []byte(thID), -1)
	if err = propPath.WriteFile(content); err != nil {
		return err
	}

	return nil
}

func restoreSketch(sk *sketch.Sketch) {
	tempFile := sk.FullPath.Join(tempPropertiesFile)
	if ex, err := tempFile.ExistCheck(); !ex || err != nil {
		return
	}
	originalFile := sk.FullPath.Join(thingPropertiesFile)
	tempFile.Rename(originalFile)
}

func uploadSketch(sk *sketch.Sketch, fqbn, port string) error {
	comm, err := cli.NewCommander()
	if err != nil {
		return err
	}

	// Compile the sketch
	err = comm.Compile(fqbn, sk.MainFile.String())
	if err != nil {
		return fmt.Errorf("cannot compile sketch: %w", err)
	}

	errMsg := "Error while uploading the custom sketch, try to put the board in bootloader mode"
	err = device.Retry(5, time.Millisecond*500, errMsg, func() error {
		serialutils.Reset(port, true, nil)
		time.Sleep(300 * time.Millisecond)
		return comm.Upload(fqbn, sk.MainFile.String(), port)
	})
	if err != nil {
		return fmt.Errorf("cannot upload sketch: %w", err)
	}

	return nil
}
