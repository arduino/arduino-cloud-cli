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

package dashboard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/template"
	"github.com/sirupsen/logrus"
)

// ExtractParams contains the parameters needed to
// extract a template dashboard from Arduino IoT Cloud and save it on local storage.
type ExtractParams struct {
	ID      string
	Format  string  // Format determines the file format of the template ("json" or "yaml")
	Outfile *string // Destination path of the extracted template
}

// Extract command is used to extract a dashboard template
// from a dashboard on Arduino IoT Cloud.
func Extract(params *ExtractParams) error {
	params.Format = strings.ToLower(params.Format)
	if params.Format != "json" && params.Format != "yaml" {
		return errors.New("format is not valid: only 'json' and 'yaml' are supported")
	}

	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	dashboard, err := iotClient.DashboardShow(params.ID)
	if err != nil {
		err = fmt.Errorf("%s: %w", "cannot extract dashboard: ", err)
		return err
	}

	templ := template.FromDashboard(dashboard)

	if params.Outfile == nil {
		name, ok := templ["name"].(string)
		if name == "" || !ok {
			return errors.New("dashboard template does not have a valid name")
		}
		name = strings.Join(strings.Fields(name), "")
		outfile := name + "-dashboard." + params.Format
		params.Outfile = &outfile
	}

	logrus.Infof("Extracting template in file: %s", *params.Outfile)
	err = template.ToFile(templ, *params.Outfile, params.Format)
	if err != nil {
		return fmt.Errorf("saving template: %w", err)
	}

	return nil
}
