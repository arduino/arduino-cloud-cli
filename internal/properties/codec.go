package properties

import (
	"github.com/bcmi-labs/oniudra-cli/iot/codec"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var senmlCodec = codec.NewSenMLCodecWithLoggerString(codec.CBOR, log)

// NewInteger creates a new senml/cbor property named `name` with the integer value `value`.
func NewInteger(name string, value int) ([]byte, error) {
	values := codec.DevicePropertyValues{
		Values: []codec.PropertyValue{},
	}
	values.AddPropertyValueNamed(name, value)

	val, err := senmlCodec.Encode(values)
	if err != nil {
		return []byte{}, err
	}

	return val, nil
}
