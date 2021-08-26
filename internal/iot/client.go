package iot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/antihax/optional"
	iotclient "github.com/arduino/iot-client-go"
)

// Client can be used to perform actions on Arduino IoT Cloud.
type Client interface {
	AddDevice(fqbn, name, serial, devType string) (string, error)
	DeleteDevice(id string) error
	ListDevices() ([]iotclient.ArduinoDevicev2, error)
	AddCertificate(id, csr string) (*iotclient.ArduinoCompressedv2, error)
	AddThing(thing *iotclient.Thing, force bool) (string, error)
	GetThing(id string) (*iotclient.ArduinoThing, error)
}

type client struct {
	ctx context.Context
	api *iotclient.APIClient
}

// NewClient returns a new client implementing the Client interface.
// It needs a ClientID and SecretID for cloud authentication.
func NewClient(clientID, secretID string) (Client, error) {
	cl := &client{}
	err := cl.setup(clientID, secretID)
	if err != nil {
		err = fmt.Errorf("instantiate new iot client: %w", err)
		return nil, err
	}
	return cl, nil
}

// AddDevice allows to create a new device on Arduino IoT Cloud.
// It returns the ID associated to the new device, and an error.
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

// DeleteDevice deletes the device corresponding to the passed ID
// from Arduino IoT Cloud.
func (cl *client) DeleteDevice(id string) error {
	_, err := cl.api.DevicesV2Api.DevicesV2Delete(cl.ctx, id)
	if err != nil {
		err = fmt.Errorf("deleting device: %w", err)
		return err
	}
	return nil
}

// ListDevices retrieves and returns a list of all Arduino IoT Cloud devices
// belonging to the user performing the request.
func (cl *client) ListDevices() ([]iotclient.ArduinoDevicev2, error) {
	devices, _, err := cl.api.DevicesV2Api.DevicesV2List(cl.ctx, nil)
	if err != nil {
		err = fmt.Errorf("listing devices: %w", err)
		return nil, err
	}
	return devices, nil
}

// AddCertifcate allows to upload a certificate on Arduino IoT Cloud.
// It returns the certificate parameters populated by the cloud.
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

// AddThing adds a new thing on Arduino IoT Cloud.
func (cl *client) AddThing(thing *iotclient.Thing, force bool) (string, error) {
	opt := &iotclient.ThingsV2CreateOpts{Force: optional.NewBool(force)}
	newThing, resp, err := cl.api.ThingsV2Api.ThingsV2Create(cl.ctx, *thing, opt)
	if err != nil {
		var respObj map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respObj)
		resp.Body.Close()
		return "", fmt.Errorf("%s: %s: %v", "adding new thing", err, respObj)
	}
	return newThing.Id, nil
}

// GetThing allows to retrieve a specific thing, given its id,
// from Arduino IoT Cloud.
func (cl *client) GetThing(id string) (*iotclient.ArduinoThing, error) {
	thing, _, err := cl.api.ThingsV2Api.ThingsV2Show(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving thing, %w", err)
		return nil, err
	}
	return &thing, nil
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
