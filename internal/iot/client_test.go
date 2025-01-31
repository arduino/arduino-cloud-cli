package iot

import (
	"testing"

	iotclient "github.com/arduino/iot-client-go/v3"
	"github.com/stretchr/testify/assert"
)

func TestJSON_UnknownFields_areAccepted(t *testing.T) {

	cert := iotclient.ArduinoDevicev2Cert{}

	// Add unknown fields to the JSON and verify that marshalling and unmarshalling works without raising error.
	// This is useful when the API is extended with new fields and the client is not updated yet.
	certJson := `{
		"compressed": {
			"not_after": "0001-01-01T00:00:00Z",
			"not_before": "0001-01-01T00:00:00Z",
			"serial": "",
			"signature": "signature",
			"signature_asn1_x": "",
			"signature_asn1_y": ""
		},
		"der": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA",
		"device_id": "123",
		"enabled": true,
		"href": "",
		"id": "",
		"pem": "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA",
		"unknown_field": "value",
		"unknown_field2": "value2",
		"new_api_field": 2222
	}`

	err := cert.UnmarshalJSON([]byte(certJson))
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %s", err)
	}
	assert.Equal(t, 3, len(cert.AdditionalProperties))
}
