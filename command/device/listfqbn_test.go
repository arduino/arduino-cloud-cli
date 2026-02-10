// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFilterFQBN(t *testing.T) {
	var (
		wrong = []FQBNInfo{
			{Name: "Arduino Uno", Value: "arduino:avr:uno", Package: "arduino"},
			{Name: "Arduino Industrial 101", Value: "arduino:avr:chiwawa", Package: "arduino"},
			{Name: "SmartEverything Lion (Native USB Port)", Value: "Arrow:samd:SmartEverything_Lion_native", Package: "Arrow"},
			{Name: "Arduino/Genuino 101", Value: "Intel:arc32:arduino_101", Package: "Intel"},
			{Name: "Atmel atmega328pb Xplained mini", Value: "atmel-avr-xminis:avr:atmega328pb_xplained_mini", Package: "atmel-avr-xminis"},
		}
		good = []FQBNInfo{
			{Name: "Arduino Nano RP2040 Connect", Value: "arduino:mbed_nano:nanorp2040connect", Package: "arduino"},
			{Name: "Arduino MKR WiFi 1010", Value: "arduino:samd:mkrwifi1010", Package: "arduino"},
			{Name: "ESP32 Dev Module", Value: "esp32:esp32:esp32", Package: "esp32"},
			{Name: "4D Systems gen4 IoD Range", Value: "esp8266:esp8266:gen4iod", Package: "esp8266"},
			{Name: "BPI-BIT", Value: "esp32:esp32:bpi-bit", Package: "esp32"},
		}
	)
	all := append(wrong, good...)
	filtered := filterFQBN(all)
	if !cmp.Equal(good, filtered) {
		t.Errorf("Wrong filter, diff:\n%s", cmp.Diff(good, filtered))
	}
}
