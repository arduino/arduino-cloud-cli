package iot

import (
	"context"
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
	UpdateThing(id string, thing *iotclient.Thing, force bool) error
	DeleteThing(id string) error
	GetThing(id string) (*iotclient.ArduinoThing, error)
	ListThings(ids []string, device *string, props bool) ([]iotclient.ArduinoThing, error)
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
		err = fmt.Errorf("creating device, %w", errorDetail(err))
		return "", err
	}
	return dev.Id, nil
}

// DeleteDevice deletes the device corresponding to the passed ID
// from Arduino IoT Cloud.
func (cl *client) DeleteDevice(id string) error {
	_, err := cl.api.DevicesV2Api.DevicesV2Delete(cl.ctx, id)
	if err != nil {
		err = fmt.Errorf("deleting device: %w", errorDetail(err))
		return err
	}
	return nil
}

// ListDevices retrieves and returns a list of all Arduino IoT Cloud devices
// belonging to the user performing the request.
func (cl *client) ListDevices() ([]iotclient.ArduinoDevicev2, error) {
	devices, _, err := cl.api.DevicesV2Api.DevicesV2List(cl.ctx, nil)
	if err != nil {
		err = fmt.Errorf("listing devices: %w", errorDetail(err))
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
		err = fmt.Errorf("creating certificate, %w", errorDetail(err))
		return nil, err
	}

	return &newCert.Compressed, nil
}

// AddThing adds a new thing on Arduino IoT Cloud.
func (cl *client) AddThing(thing *iotclient.Thing, force bool) (string, error) {
	opt := &iotclient.ThingsV2CreateOpts{Force: optional.NewBool(force)}
	newThing, _, err := cl.api.ThingsV2Api.ThingsV2Create(cl.ctx, *thing, opt)
	if err != nil {
		return "", fmt.Errorf("%s: %w", "adding new thing", errorDetail(err))
	}
	return newThing.Id, nil
}

// AddThing updates a thing on Arduino IoT Cloud.
func (cl *client) UpdateThing(id string, thing *iotclient.Thing, force bool) error {
	opt := &iotclient.ThingsV2UpdateOpts{Force: optional.NewBool(force)}
	_, _, err := cl.api.ThingsV2Api.ThingsV2Update(cl.ctx, id, *thing, opt)
	if err != nil {
		return fmt.Errorf("%s: %v", "updating thing", errorDetail(err))
	}
	return nil
}

// DeleteThing deletes a thing from Arduino IoT Cloud.
func (cl *client) DeleteThing(id string) error {
	_, err := cl.api.ThingsV2Api.ThingsV2Delete(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting thing: %w", errorDetail(err))
		return err
	}
	return nil
}

// GetThing allows to retrieve a specific thing, given its id,
// from Arduino IoT Cloud.
func (cl *client) GetThing(id string) (*iotclient.ArduinoThing, error) {
	thing, _, err := cl.api.ThingsV2Api.ThingsV2Show(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving thing, %w", errorDetail(err))
		return nil, err
	}
	return &thing, nil
}

// ListThings returns a list of things on Arduino IoT Cloud.
func (cl *client) ListThings(ids []string, device *string, props bool) ([]iotclient.ArduinoThing, error) {
	opts := &iotclient.ThingsV2ListOpts{}
	opts.ShowProperties = optional.NewBool(props)

	if ids != nil {
		opts.Ids = optional.NewInterface(ids)
	}

	if device != nil {
		opts.DeviceId = optional.NewString(*device)
	}

	things, _, err := cl.api.ThingsV2Api.ThingsV2List(cl.ctx, opts)
	if err != nil {
		err = fmt.Errorf("retrieving things, %w", errorDetail(err))
		return nil, err
	}
	return things, nil
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
