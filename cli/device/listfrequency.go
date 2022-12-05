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

package device

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListFrequencyPlansCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list-frequency-plans",
		Short: "List LoRa frequency plans",
		Long:  "List all supported LoRa frequency plans",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListFrequencyPlansCommand(); err != nil {
				feedback.Errorf("Error during device list-frequency-plans: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	return listCommand
}

func runListFrequencyPlansCommand() error {
	logrus.Info("Listing supported frequency plans")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	freqs, err := device.ListFrequencyPlans(context.TODO(), cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(listFrequencyPlansResult{freqs})
	return nil
}

type listFrequencyPlansResult struct {
	freqs []device.FrequencyPlanInfo
}

func (r listFrequencyPlansResult) Data() interface{} {
	return r.freqs
}

func (r listFrequencyPlansResult) String() string {
	if len(r.freqs) == 0 {
		return "No frequency plan found."
	}
	t := table.New()
	t.SetHeader("ID", "Name", "Advanced")
	for _, freq := range r.freqs {
		t.AddRow(
			freq.ID,
			freq.Name,
			freq.Advanced,
		)
	}
	return t.Render()
}
