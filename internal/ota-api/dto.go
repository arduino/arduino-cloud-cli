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

package otaapi

import (
	"strconv"
	"strings"
	"time"

	"unicode"

	"github.com/arduino/arduino-cli/table"
)

const progressBarMultiplier = 2

type (
	OtaStatusResponse struct {
		Ota    Ota     `json:"ota"`
		States []State `json:"states,omitempty"`
	}

	OtaStatusList struct {
		Ota []Ota `json:"ota"`
	}

	Ota struct {
		ID           string `json:"id,omitempty" yaml:"id,omitempty"`
		DeviceID     string `json:"device_id,omitempty" yaml:"device_id,omitempty"`
		Status       string `json:"status" yaml:"status"`
		StartedAt    string `json:"started_at" yaml:"started_at"`
		EndedAt      string `json:"ended_at,omitempty" yaml:"ended_at,omitempty"`
		ErrorReason  string `json:"error_reason,omitempty" yaml:"error_reason,omitempty"`
		FirmwareSize int64  `json:"firmware_size,omitempty"`
		MaxRetries   int64  `json:"max_retries,omitempty"`
		RetryAttempt int64  `json:"retry_attempt,omitempty"`
	}

	State struct {
		OtaID     string `json:"ota_id"`
		State     string `json:"state"`
		StateData string `json:"state_data,omitempty"`
		Timestamp string `json:"timestamp,omitempty"`
	}

	OtaStatusDetail struct {
		FirmwareSize int64   `json:"firmware_size,omitempty"`
		MaxRetries   int64   `json:"max_retries,omitempty"`
		RetryAttempt int64   `json:"retry_attempt,omitempty"`
		Ota          Ota     `json:"ota"`
		Details      []State `json:"details,omitempty"`
	}
)

func (r OtaStatusList) Data() interface{} {
	return r.Ota
}

func (r OtaStatusList) String() string {
	if len(r.Ota) == 0 {
		return ""
	}
	t := table.New()
	hasErrorReason := false
	for _, r := range r.Ota {
		if r.ErrorReason != "" {
			hasErrorReason = true
			break
		}
	}

	if hasErrorReason {
		t.SetHeader("Device ID", "Ota ID", "Status", "Started At", "Ended At", "Error Reason", "Retry Attempt")
	} else {
		t.SetHeader("Device ID", "Ota ID", "Status", "Started At", "Ended At")
	}

	// Now print the table
	for _, r := range r.Ota {
		line := []any{r.DeviceID, r.ID, r.MapStatus(), formatHumanReadableTs(r.StartedAt), formatHumanReadableTs(r.EndedAt)}
		if hasErrorReason {
			line = append(line, r.ErrorReason)
			line = append(line, strconv.FormatInt(r.RetryAttempt, 10))
		}
		t.AddRow(line...)
	}

	return t.Render()
}

func (o Ota) MapStatus() string {
	return upperCaseFirst(o.Status)
}

func (r Ota) Data() interface{} {
	return r
}

func (r Ota) String() string {
	if len(r.ID) == 0 {
		return ""
	}
	t := table.New()
	hasErrorReason := r.ErrorReason != ""

	if hasErrorReason {
		t.SetHeader("Device ID", "Ota ID", "Status", "Started At", "Ended At", "Error Reason", "Retry Attempt")
	} else {
		t.SetHeader("Device ID", "Ota ID", "Status", "Started At", "Ended At")
	}

	// Now print the table
	line := []any{r.DeviceID, r.ID, r.MapStatus(), formatHumanReadableTs(r.StartedAt), formatHumanReadableTs(r.EndedAt)}
	if hasErrorReason {
		line = append(line, r.ErrorReason)
		line = append(line, strconv.FormatInt(r.RetryAttempt, 10))
	}
	t.AddRow(line...)

	return t.Render()
}

func (r OtaStatusDetail) Data() interface{} {
	return r.Ota
}

func (r OtaStatusDetail) String() string {
	if r.Ota.ID == "" {
		return "No OTA found"
	}
	t := table.New()

	succeeded := strings.ToLower(r.Ota.Status) == "succeeded"
	hasError := r.Ota.ErrorReason != "" || !succeeded

	if hasError {
		t.SetHeader("Device ID", "Ota ID", "Status", "Started At", "Ended At", "Error Reason", "Retry Attempt")
	} else {
		t.SetHeader("Device ID", "Ota ID", "Status", "Started At", "Ended At")
	}

	// Now print the table
	line := []any{r.Ota.DeviceID, r.Ota.ID, r.Ota.MapStatus(), formatHumanReadableTs(r.Ota.StartedAt), formatHumanReadableTs(r.Ota.EndedAt)}
	if hasError {
		line = append(line, r.Ota.ErrorReason)
		line = append(line, strconv.FormatInt(r.RetryAttempt, 10))
	}
	t.AddRow(line...)

	output := t.Render()

	// Add details
	if len(r.Details) > 0 {
		t = table.New()
		t.SetHeader("Time", "Status", "Detail")
		fwSize := int64(0)
		if r.FirmwareSize > 0 {
			fwSize = r.FirmwareSize
		}

		firstTS := formatHumanReadableTs(r.Details[0].Timestamp)
		if !containsResetState(r.Details) && !hasError {
			t.AddRow(firstTS, "Flash", "")
		}

		hasReachedFlashState := hasReachedFlashState(r.Details, succeeded)
		for _, s := range r.Details {
			stateData := formatStateData(s.State, s.StateData, fwSize, hasReachedFlashState)
			t.AddRow(formatHumanReadableTs(s.Timestamp), upperCaseFirst(s.State), stateData)
		}

		output += "\nDetails:\n" + t.Render()
	}

	return output
}

func hasReachedFlashState(states []State, succeeded bool) bool {
	if succeeded {
		return true
	}
	return containsResetState(states)
}

func containsResetState(states []State) bool {
	for _, s := range states {
		if strings.ToLower(s.State) == "flash" || strings.ToLower(s.State) == "reboot" {
			return true
		}
	}
	return false
}

func formatStateData(state, data string, firmware_size int64, hasReceivedFlashState bool) string {
	if data == "" || data == "Unknown" {
		return ""
	}
	if state == "fetch" {
		// This is the state 'fetch' of OTA progress. This contains a number that represents the number of bytes fetched
		actualDownloadedData, err := strconv.Atoi(data)
		if err != nil || actualDownloadedData <= 0 || firmware_size <= 0 { // Sanitize and avoid division by zero
			return data
		}
		if hasReceivedFlashState {
			return buildSimpleProgressBar(float64(100), firmware_size)
		}
		percentage := (float64(actualDownloadedData) / float64(firmware_size)) * 100
		return buildSimpleProgressBar(percentage, firmware_size)
	}
	return data
}

func buildSimpleProgressBar(progress float64, fw_size int64) string {
	progressInt := int(progress) / 10
	progressInt = progressInt * progressBarMultiplier
	maxProgress := 10 * progressBarMultiplier
	var bar strings.Builder
	bar.WriteString("[")
	bar.WriteString(strings.Repeat("=", progressInt))
	bar.WriteString(strings.Repeat(" ", maxProgress-progressInt))
	bar.WriteString("] ")
	bar.WriteString(strconv.FormatFloat(progress, 'f', 0, 64))
	bar.WriteString("% (firmware size: ")
	bar.WriteString(strconv.FormatInt(fw_size, 10))
	bar.WriteString(" bytes)")
	return bar.String()
}

func upperCaseFirst(s string) string {
	if len(s) > 0 {
		s = strings.ReplaceAll(s, "_", " ")
		for i, v := range s {
			return string(unicode.ToUpper(v)) + s[i+1:]
		}
	}
	return ""
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
