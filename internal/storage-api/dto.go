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

package storageapi

import (
	"time"

	"github.com/arduino/arduino-cli/table"
)

type (
	ImportCustomTemplateResponse struct {
		Message    string `json:"message"`
		Name       string `json:"name"`
		TemplateId string `json:"template_id"`
	}
	TemplateEntry struct {
		TemplateId string `json:"template_id"`
		Name       string `json:"name"`
		CreatedAt  string `json:"created_at"`
	}
	TemplatesListResponse struct {
		Templates []TemplateEntry `json:"templates"`
	}
)

func (r *TemplatesListResponse) Data() interface{} {
	return r.Templates
}

func (r *TemplatesListResponse) String() string {
	if len(r.Templates) == 0 {
		return ""
	}
	t := table.New()
	t.SetHeader("Template ID", "Name", "Created At")

	// Now print the table
	for _, tem := range r.Templates {
		line := []any{tem.TemplateId, tem.Name, formatHumanReadableTs(tem.CreatedAt)}
		t.AddRow(line...)
	}

	return t.Render()
}

func formatHumanReadableTs(ts string) string {
	if ts == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return parsed.Format(time.RFC3339)
}
