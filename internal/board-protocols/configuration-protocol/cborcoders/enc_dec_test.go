// This file is part of arduino-cloud-cli.
//
// Copyright ARDUINO SRL http://www.arduino.cc/)
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

package cborcoders

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecode(t *testing.T) {
	fixedTime := uint64(1709208245) // time.Date(2024, 2, 29, 12, 4, 5, 0, time.UTC)

	tests := []struct {
		name string
		in   Cmd
		want string
	}{
		{
			name: "provisioning status",
			in:   From(ProvisioningStatusMessage{Status: -100}),
			want: "da00012000813863",
		},
		{
			name: "provisioning BLE mac address",
			in: From(ProvisioningBLEMacAddressMessage{
				BLEMacAddress: [6]uint8{0xAF, 0xAF, 0xAF, 0xAF, 0xAF, 0xAF}}),
			want: "DA000120138146AFAFAFAFAFAF",
		},
		{
			name: "provisioning WiFi FW Version",
			in: From(ProvisioningWiFiFWVersionMessage{
				WiFiFWVersion: "1.6.0"}),
			want: "DA000120148165312E362E30",
		},
		{
			name: "provisioning sketch Version",
			in: From(ProvisioningSketchVersionMessage{
				ProvisioningSketchVersion: "1.6.0"}),
			want: "DA000120158165312E362E30",
		},
		{
			name: "provisioning network configurator Version",
			in: From(ProvisioningNetworkConfigLibVersionMessage{
				NetworkConfigLibVersion: "1.6.0"}),
			want: "DA000120168165312E362E30",
		},
		{
			name: "provisioning unique id",
			in: From(ProvisioningUniqueIdMessage{
				UniqueId: [32]uint8{0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA}}),
			want: "DA00012010815820CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA",
		},
		{
			name: "provisioning signature",
			in: From(ProvisioningSignatureMessage{
				Signature: [268]uint8{0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA, 0xCA,
					0xCA, 0xCA, 0xCA, 0xCA}}),
			want: "DA000120118159010CCACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"CACACACACACACACACACACACACACACACACACACACACACACACACACACACACACA" +
				"cacacacacacacacacacacacacacacacacacacacacacacacacacacacacaca" +
				"cacacacacacacacacaca",
		},
		{
			name: "provisioning public key",
			in: From(ProvisioningPublicKeyMessage{
				ProvisioningPublicKey: "-----BEGIN PUBLIC KEY-----\n" +
					"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7JxCtXl5SvIrHmiasqyN4pyoXRlm44d5WXNpqmvJ\n" +
					"k0tH8UpmIeHG7YPAkKLaqid95v/wLVoWeX5EbjxmlCkFtw==\n-----END PUBLIC KEY-----\n",
			}),
			want: "DA000120178178B22D2D2D2D2D424547494E205055424C4943204B45592D2D2D2D2" +
				"D0A4D466B77457759484B6F5A497A6A3043415159494B6F5A497A6A304441516344" +
				"51674145374A784374586C3553764972486D69617371794E3470796F58526C6D343" +
				"4643557584E70716D764A0A6B3074483855706D49654847375950416B4B4C617169" +
				"643935762F774C566F5765583545626A786D6C436B4674773D3D0A2D2D2D2D2D454" +
				"E44205055424C4943204B45592D2D2D2D2D0A",
		},
		{
			name: "provisioning timestamp",
			in:   From(ProvisioningTimestampMessage{Timestamp: fixedTime}),
			want: "DA00012002811A65E072B5",
		},
		{
			name: "provisioning commands",
			in:   From(ProvisioningCommandsMessage{Command: 100}),
			want: "DA00012003811864",
		},
		{
			name: "provisioning wifi config",
			in:   From(ProvisioningWifiConfigMessage{SSID: "SSID1", PWD: "PASSWORDSSID1"}),
			want: "DA00012004826553534944316D50415353574F52445353494431",
		},
		{
			name: "provisioning lora config",
			in: From(ProvisioningLoRaConfigMessage{
				AppEui:      "APPEUI1",
				AppKey:      "APPKEY",
				Band:        5,
				ChannelMask: "01110",
				DeviceClass: "A",
			}),
			want: "DA00012005856741505045554931664150504B4559056530313131306141",
		},
		{
			name: "provisioning gsm config",
			in: From(ProvisioningGSMConfigMessage{
				PIN:   "12345678",
				Apn:   "apn.arduino.cc",
				Login: "TESTUSER",
				Pass:  "TESTPASSWORD",
			}),
			want: "DA00012006846831323334353637386E61706E2E61726475696E6F2E63636854455354555345526C5445535450415353574F5244",
		},
		{
			name: "provisoning gsm config without pin",
			in: From(ProvisioningGSMConfigMessage{
				Apn:   "apn.arduino.cc",
				Login: "TESTUSER",
				Pass:  "TESTPASSWORD",
			}),
			want: "DA0001200684606E61706E2E61726475696E6F2E63636854455354555345526C5445535450415353574F5244",
		},
		{
			name: "provisioning nb config",
			in: From(ProvisioningNBConfigMessage{
				PIN:   "12345678",
				Apn:   "apn.arduino.cc",
				Login: "TESTUSER",
				Pass:  "TESTPASSWORD",
			}),
			want: "DA00012007846831323334353637386E61706E2E61726475696E6F2E63636854455354555345526C5445535450415353574F5244",
		},
		{
			name: "provisioning nb config without pin, login and pass",
			in: From(ProvisioningNBConfigMessage{
				PIN:   "",
				Apn:   "apn.arduino.cc",
				Login: "",
				Pass:  "",
			}),
			want: "DA0001200784606E61706E2E61726475696E6F2E63636060",
		},
		{
			name: "provisioning catm1 config",
			in: From(ProvisioningCATM1ConfigMessage{
				PIN:   "12345678",
				Band:  []uint32{1, 2, 524288, 134217728},
				Apn:   "apn.arduino.cc",
				Login: "TESTUSER",
				Pass:  "TESTPASSWORD",
			}),
			want: "DA00012008856831323334353637388401021A000800001A080000006E61706E2E61726475696E6F2E63636854455354555345526C5445535450415353574F5244",
		},
		{
			name: "provisioning catm1 config no band",
			in: From(ProvisioningCATM1ConfigMessage{
				PIN:   "12345678",
				Band:  []uint32{},
				Apn:   "apn.arduino.cc",
				Login: "TESTUSER",
				Pass:  "TESTPASSWORD",
			}),
			want: "DA0001200885683132333435363738806E61706E2E61726475696E6F2E63636854455354555345526C5445535450415353574F5244",
		},
		{
			name: "provisioning ethernet config ipv4",
			in: From(ProvisioningEthernetConfigMessage{
				Static_ip:       []byte{192, 168, 0, 2},
				Dns:             []byte{8, 8, 8, 8},
				Gateway:         []byte{192, 168, 1, 1},
				Netmask:         []byte{255, 255, 255, 0},
				Timeout:         15,
				ResponseTimeout: 200,
			}),
			want: "DA000120098644C0A80002440808080844C0A8010144FFFFFF000F18C8",
		},
		{
			name: "provisioning ethernet config ipv6",
			in: From(ProvisioningEthernetConfigMessage{
				Static_ip:       []byte{0x1a, 0x4f, 0xa7, 0xa9, 0x92, 0x8f, 0x7b, 0x1c, 0xec, 0x3b, 0x1e, 0xcd, 0x88, 0x58, 0x0d, 0x1e},
				Dns:             []byte{0x21, 0xf6, 0x3b, 0x22, 0x99, 0x6f, 0x5b, 0x72, 0x25, 0xd9, 0xe0, 0x24, 0xf0, 0x36, 0xb5, 0xd2},
				Gateway:         []byte{0x2e, 0xc2, 0x27, 0xf1, 0xf1, 0x9a, 0x0c, 0x11, 0x47, 0x1b, 0x84, 0xaf, 0x96, 0x10, 0xb0, 0x17},
				Netmask:         []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				Timeout:         15,
				ResponseTimeout: 200,
			}),
			want: "DA0001200986501A4FA7A9928F7B1CEC3B1ECD88580D1E5021F63B22996F5B7225D9E024F036B5D2502EC227F1F19A0C11471B84AF9610B01750FFFFFFFFFFFFFFFF00000000000000000F18C8",
		},
		{
			name: "provisioning cellular config",
			in: From(ProvisioningCellularConfigMessage{
				PIN:   "12345678",
				Apn:   "apn.arduino.cc",
				Login: "TESTUSER",
				Pass:  "TESTPASSWORD",
			}),
			want: "DA00012012846831323334353637386E61706E2E61726475696E6F2E63636854455354555345526C5445535450415353574F5244",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.in.Encode()
			assert.NoError(t, err)
			tt.want = strings.ToLower(tt.want)
			hexGot := hex.EncodeToString(got)
			assert.Equal(t, tt.want, hexGot)

			cmd, err := Decode(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.in, cmd)
		})
	}
}

func TestDecodeWiFiList(t *testing.T) {
	list := From(WiFiNetworks{{SSID: "SSID1", RSSI: -76}, {SSID: "SSID2", RSSI: -56}})
	encoded_list, _ := hex.DecodeString("DA0001200184655353494431384B6553534944323837")
	decoded, err := Decode(encoded_list)
	assert.NoError(t, err)
	assert.Equal(t, list, decoded)
}
