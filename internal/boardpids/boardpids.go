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

package boardpids

var (
	BoardTypes = map[uint32]string{
		0x45535033: "ESP32",
		0x23418054: "MKR_WIFI_1010",
		0x23418057: "NANO_33_IOT",
		0x2341025B: "PORTENTA_H7_M7",
		0x2341005E: "NANO_RP2040_CONNECT",
		0x2341025F: "NICLA_VISION",
		0x23410064: "OPTA",
		0x23410266: "GIGA",
		0x23410070: "NANO_ESP32",
		0x23411002: "UNOR4WIFI",
	}

	ArduinoPidToFQBN = map[string]string{
		"1002": "arduino:renesas_uno:unor4wifi",
		"0070": "arduino:esp32:nano_nora",
		"8057": "arduino:samd:nano_33_iot",
		"804E": "arduino:samd:mkr1000",
		"8052": "arduino:samd:mkrgsm1400",
		"8055": "arduino:samd:mkrnb1500",
		"8054": "arduino:samd:mkrwifi1010",
		"005E": "arduino:mbed_nano:nanorp2040connect",
		"025B": "arduino:mbed_portenta:envie_m7",
		"025F": "arduino:mbed_nicla:nicla_vision",
		"0064": "arduino:mbed_opta:opta",
		"0266": "arduino:mbed_giga:giga",
		"0068": "arduino:renesas_portenta:portenta_c33",
	}

	ArduinoFqbnToPID = map[string]string{
		"arduino:samd:nano_33_iot":              "8057",
		"arduino:samd:mkr1000":                  "804E",
		"arduino:samd:mkrgsm1400":               "8052",
		"arduino:samd:mkrnb1500":                "8055",
		"arduino:samd:mkrwifi1010":              "8054",
		"arduino:mbed_nano:nanorp2040connect":   "005E",
		"arduino:mbed_portenta:envie_m7":        "025B",
		"arduino:mbed_nicla:nicla_vision":       "025F",
		"arduino:mbed_opta:opta":                "0064",
		"arduino:mbed_giga:giga":                "0266",
		"arduino:renesas_uno:unor4wifi":         "1002",
		"arduino:esp32:nano_nora":               "0070",
		"arduino:renesas_portenta:portenta_c33": "0068",
	}

	ArduinoVendorID = "2341"

	Esp32MagicNumberPart1 = "4553"
	Esp32MagicNumberPart2 = "5033"
)
