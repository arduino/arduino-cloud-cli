package cborcoders

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/fxamacker/cbor/v2"
)

var _dm cbor.DecMode
var _em cbor.EncMode

var wifiListBeginMessage = []byte{0xda, 0x00, 0x01, 0x20, 0x01}

// Provisioning commands
var ProvisioningStatusMessageType = reflect.TypeOf(ProvisioningStatusMessage{})
var WiFiNetworksType = reflect.TypeOf(WiFiNetworks{})
var ProvisioningUniqueIdMessageType = reflect.TypeOf(ProvisioningUniqueIdMessage{})
var ProvisioningBLEMacAddressMessageType = reflect.TypeOf(ProvisioningBLEMacAddressMessage{})
var ProvisioningWiFiFWVersionMessageType = reflect.TypeOf(ProvisioningWiFiFWVersionMessage{})
var ProvisioningSketchVersionMessageType = reflect.TypeOf(ProvisioningSketchVersionMessage{})
var ProvisioningNetConfigLibVersionMessageType = reflect.TypeOf(ProvisioningNetworkConfigLibVersionMessage{})
var ProvisioningSignatureMessageType = reflect.TypeOf(ProvisioningSignatureMessage{})
var ProvisioningPublicKeyMessageType = reflect.TypeOf(ProvisioningPublicKeyMessage{})
var ProvisioningTimestampMessageType = reflect.TypeOf(ProvisioningTimestampMessage{})
var ProvisioningCommandsMessageType = reflect.TypeOf(ProvisioningCommandsMessage{})
var ProvisioningWifiConfigMessageType = reflect.TypeOf(ProvisioningWifiConfigMessage{})
var ProvisioningLoRaConfigMessageType = reflect.TypeOf(ProvisioningLoRaConfigMessage{})
var ProvisioningGSMConfigMessageType = reflect.TypeOf(ProvisioningGSMConfigMessage{})
var ProvisioningNBIoTConfigMessageType = reflect.TypeOf(ProvisioningNBConfigMessage{})
var ProvisioningCATM1ConfigMessageType = reflect.TypeOf(ProvisioningCATM1ConfigMessage{})
var ProvisioningEthernetConfigMessageType = reflect.TypeOf(ProvisioningEthernetConfigMessage{})
var ProvisioningCellularConfigMessageType = reflect.TypeOf(ProvisioningCellularConfigMessage{})

type tag struct {
	tag uint64
	ty  reflect.Type
}

var tagCommands = []tag{
	// provisioning commands
	{0x012000, ProvisioningStatusMessageType},
	{0x012001, WiFiNetworksType},
	{0x012013, ProvisioningBLEMacAddressMessageType},
	{0x012014, ProvisioningWiFiFWVersionMessageType},
	{0x012015, ProvisioningSketchVersionMessageType},
	{0x012016, ProvisioningNetConfigLibVersionMessageType},
	{0x012010, ProvisioningUniqueIdMessageType},
	{0x012011, ProvisioningSignatureMessageType},
	{0x012017, ProvisioningPublicKeyMessageType},
	{0x012002, ProvisioningTimestampMessageType},
	{0x012003, ProvisioningCommandsMessageType},
	{0x012004, ProvisioningWifiConfigMessageType},
	{0x012005, ProvisioningLoRaConfigMessageType},
	{0x012006, ProvisioningGSMConfigMessageType},
	{0x012007, ProvisioningNBIoTConfigMessageType},
	{0x012008, ProvisioningCATM1ConfigMessageType},
	{0x012009, ProvisioningEthernetConfigMessageType},
	{0x012012, ProvisioningCellularConfigMessageType},
}

func init() {
	tags := cbor.NewTagSet()
	for _, t := range tagCommands {
		err := tags.Add(
			cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
			t.ty,
			t.tag)
		if err != nil {
			panic(err)
		}
	}
	var err error
	_dm, err = cbor.DecOptions{}.DecModeWithTags(tags)
	if err != nil {
		panic(err)
	}

	_em, err = cbor.EncOptions{IndefLength: 1}.EncModeWithTags(tags)

	if err != nil {
		panic(err)
	}
}

type Cmd struct {
	inner interface{}
}

func getCBORType(data byte) byte {
	return data & 0xe0
}

func getCBORFieldLength(data []byte) (len, i int) {
	additional_information := data[0] & 0x1f
	if additional_information <= 23 {
		return int(additional_information), 0
	} else if additional_information == 24 {
		return int(data[1]), 1
	} else if additional_information == 25 {
		return int(data[1])<<8 | int(data[2]), 2
	}
	return 0, 0
}

func getString(data []byte, len int) string {
	return string(data[:len])
}

func DecodeWiFiNetworks(data []byte, wf *WiFiNetworks) error {
	len := len(data)
	if len == 0 {
		return errors.New("empty data")
	}
	networks := []WiFiNetwork{}
	i := 0
	for i < len {
		if getCBORType(data[i]) != 0x80 { //check if it's an array
			return errors.New("invalid data not initial array")
		}

		array_length, l := getCBORFieldLength(data[i:])
		i += l
		i++
		for j := 0; j < array_length; j = j + 2 {
			if getCBORType(data[i]) != 0x60 { //check if it's a text string
				return errors.New("invalid data not text string")
			}

			text_length, l := getCBORFieldLength(data[i:])
			i += l
			i++
			ssid := getString(data[i:], text_length)
			i += text_length
			if getCBORType(data[i]) != 0x20 { //check if it's a negative number
				return errors.New("invalid data not negative number")
			}
			val, l := getCBORFieldLength(data[i:])
			rssi := int(-1) ^ int(val)
			i += l
			i++
			networks = append(networks, WiFiNetwork{SSID: ssid, RSSI: rssi})
		}
	}
	*wf = networks

	return nil
}

func Decode(message []byte) (cmd Cmd, err error) {
	c := Cmd{}
	if bytes.Equal(message[0:5], wifiListBeginMessage) {
		wf := WiFiNetworks{}
		e := DecodeWiFiNetworks(message[5:], &wf)
		c.inner = wf
		return c, e
	}

	if err := _dm.Unmarshal(message, &c.inner); err != nil {
		return Cmd{}, err
	}

	match := slices.ContainsFunc(tagCommands, func(t tag) bool {
		return t.ty == c.Type()
	})
	if !match {
		return Cmd{}, fmt.Errorf("unknown command type: %v", c.Type())
	}

	return c, nil

}

func (c Cmd) Encode() ([]byte, error) {
	p, err := _em.Marshal(c.inner)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c Cmd) String() string {
	idx := slices.IndexFunc(tagCommands, func(t tag) bool {
		return t.ty == c.Type()
	})
	return fmt.Sprintf("%x=%+v", tagCommands[idx].tag, c.inner)
}

func From[T ProvisioningStatusMessage | WiFiNetworks | ProvisioningBLEMacAddressMessage | ProvisioningWiFiFWVersionMessage |
	ProvisioningSketchVersionMessage | ProvisioningNetworkConfigLibVersionMessage | ProvisioningUniqueIdMessage |
	ProvisioningSignatureMessage | ProvisioningPublicKeyMessage | ProvisioningTimestampMessage |
	ProvisioningCommandsMessage | ProvisioningWifiConfigMessage |
	ProvisioningLoRaConfigMessage | ProvisioningCellularConfigMessage |
	ProvisioningEthernetConfigMessage | ProvisioningCATM1ConfigMessage |
	ProvisioningGSMConfigMessage | ProvisioningNBConfigMessage](c T) Cmd {
	return Cmd{inner: c}
}

func (c Cmd) Type() reflect.Type {
	return reflect.TypeOf(c.inner)
}

func (c Cmd) ToProvisioningStatusMessage() ProvisioningStatusMessage {
	return c.inner.(ProvisioningStatusMessage)
}

func (c Cmd) ToWiFiNetworks() WiFiNetworks {
	return c.inner.(WiFiNetworks)
}

func (c Cmd) ToProvisioningBLEMacAddressMessage() ProvisioningBLEMacAddressMessage {
	return c.inner.(ProvisioningBLEMacAddressMessage)
}

func (c Cmd) ToProvisioningWiFiFWVersionMessage() ProvisioningWiFiFWVersionMessage {
	return c.inner.(ProvisioningWiFiFWVersionMessage)
}

func (c Cmd) ToProvisioningSketchVersionMessage() ProvisioningSketchVersionMessage {
	return c.inner.(ProvisioningSketchVersionMessage)
}

func (c Cmd) ToProvisioningNetworkConfigLibVersionMessage() ProvisioningNetworkConfigLibVersionMessage {
	return c.inner.(ProvisioningNetworkConfigLibVersionMessage)
}

func (c Cmd) ToProvisioningUniqueIdMessage() ProvisioningUniqueIdMessage {
	return c.inner.(ProvisioningUniqueIdMessage)
}

func (c Cmd) ToProvisioningSignatureMessage() ProvisioningSignatureMessage {
	return c.inner.(ProvisioningSignatureMessage)
}

func (c Cmd) ToProvisioningPublicKeyMessage() ProvisioningPublicKeyMessage {
	return c.inner.(ProvisioningPublicKeyMessage)
}

func (c Cmd) ToProvisioningTimestampMessage() ProvisioningTimestampMessage {
	return c.inner.(ProvisioningTimestampMessage)
}

func (c Cmd) ToProvisioningCommandsMessage() ProvisioningCommandsMessage {
	return c.inner.(ProvisioningCommandsMessage)
}

func (c Cmd) ToProvisioningWifiConfigMessage() ProvisioningWifiConfigMessage {
	return c.inner.(ProvisioningWifiConfigMessage)
}

func (c Cmd) ToProvisioningLoRaConfigMessage() ProvisioningLoRaConfigMessage {
	return c.inner.(ProvisioningLoRaConfigMessage)
}

func (c Cmd) ToProvisioningCellularConfigMessage() ProvisioningCellularConfigMessage {
	return c.inner.(ProvisioningCellularConfigMessage)
}

func (c Cmd) ToProvisioningEthernetConfigMessage() ProvisioningEthernetConfigMessage {
	return c.inner.(ProvisioningEthernetConfigMessage)
}

func (c Cmd) ToProvisioningCATM1ConfigMessage() ProvisioningCATM1ConfigMessage {
	return c.inner.(ProvisioningCATM1ConfigMessage)
}

func (c Cmd) ToProvisioningGSMConfigMessage() ProvisioningGSMConfigMessage {
	return c.inner.(ProvisioningGSMConfigMessage)
}

func (c Cmd) ToProvisioningNBConfigMessage() ProvisioningNBConfigMessage {
	return c.inner.(ProvisioningNBConfigMessage)
}
