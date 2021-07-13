package iot

import (
	"context"
	"fmt"

	iotclient "github.com/arduino/iot-client-go"
)

type Client interface {
	AddDevice(fqbn, name, serial, devType string) (string, error)
	AddCertificate(id, csr string) (*iotclient.ArduinoCompressedv2, error)
}

type client struct {
	ctx context.Context
	api *iotclient.APIClient
}

func NewClient(clientID, secretID string) (Client, error) {
	cl := &client{}
	err := cl.setup(clientID, secretID)
	if err != nil {
		err = fmt.Errorf("instantiate new iot client: %w", err)
		return nil, err
	}
	return cl, nil
}

func (cl *client) AddDevice(fqbn, name, serial, dType string) (string, error) {
	payload := iotclient.CreateDevicesV2Payload{
		Fqbn:   fqbn,
		Name:   name,
		Serial: serial,
		Type:   dType,
	}
	dev, _, err := cl.api.DevicesV2Api.DevicesV2Create(cl.ctx, payload)
	if err != nil {
		err = fmt.Errorf("creating device, %w", err)
		return "", err
	}
	return dev.Id, nil
}

func (cl *client) AddCertificate(id, csr string) (*iotclient.ArduinoCompressedv2, error) {
	cert := iotclient.CreateDevicesV2CertsPayload{
		Ca:      "Arduino",
		Csr:     csr,
		Enabled: true,
	}

	newCert, _, err := cl.api.DevicesV2CertsApi.DevicesV2CertsCreate(cl.ctx, id, cert)
	if err != nil {
		err = fmt.Errorf("creating certificate, %w", err)
		return nil, err
	}

	return &newCert.Compressed, nil
}

func (cl *client) setup(client, secret string) error {
	// Get the access token in exchange of client_id and client_secret
	tok, err := token(client, secret)
	if err != nil {
		err = fmt.Errorf("cannot retrieve token given client and secret: %w", err)
		return err
	}

	// We use the token to create a context that will be passed to any API call
	cl.ctx = context.WithValue(context.Background(), iotclient.ContextAccessToken, tok.AccessToken)

	// Create an instance of the iot-api Go client, we pass an empty config
	// because defaults are ok
	cl.api = iotclient.NewAPIClient(iotclient.NewConfiguration())

	return nil
}
