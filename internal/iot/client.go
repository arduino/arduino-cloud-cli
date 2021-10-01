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

package iot

import (
	"context"
	"fmt"
	"os"

	"github.com/antihax/optional"
	iotclient "github.com/arduino/iot-client-go"
)

// Client can be used to perform actions on Arduino IoT Cloud.
type Client interface {
	DeviceCreate(fqbn, name, serial, devType string) (*iotclient.ArduinoDevicev2, error)
	DeviceDelete(id string) error
	DeviceList() ([]iotclient.ArduinoDevicev2, error)
	DeviceShow(id string) (*iotclient.ArduinoDevicev2, error)
	DeviceOTA(id string, file *os.File, expireMins int) error
	CertificateCreate(id, csr string) (*iotclient.ArduinoCompressedv2, error)
	ThingCreate(thing *iotclient.Thing, force bool) (*iotclient.ArduinoThing, error)
	ThingUpdate(id string, thing *iotclient.Thing, force bool) error
	ThingDelete(id string) error
	ThingShow(id string) (*iotclient.ArduinoThing, error)
	ThingList(ids []string, device *string, props bool) ([]iotclient.ArduinoThing, error)
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

// DeviceCreate allows to create a new device on Arduino IoT Cloud.
// It returns the newly created device, and an error.
func (cl *client) DeviceCreate(fqbn, name, serial, dType string) (*iotclient.ArduinoDevicev2, error) {
	payload := iotclient.CreateDevicesV2Payload{
		Fqbn:   fqbn,
		Name:   name,
		Serial: serial,
		Type:   dType,
	}
	dev, _, err := cl.api.DevicesV2Api.DevicesV2Create(cl.ctx, payload)
	if err != nil {
		err = fmt.Errorf("creating device, %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DeviceDelete deletes the device corresponding to the passed ID
// from Arduino IoT Cloud.
func (cl *client) DeviceDelete(id string) error {
	_, err := cl.api.DevicesV2Api.DevicesV2Delete(cl.ctx, id)
	if err != nil {
		err = fmt.Errorf("deleting device: %w", errorDetail(err))
		return err
	}
	return nil
}

// DeviceList retrieves and returns a list of all Arduino IoT Cloud devices
// belonging to the user performing the request.
func (cl *client) DeviceList() ([]iotclient.ArduinoDevicev2, error) {
	devices, _, err := cl.api.DevicesV2Api.DevicesV2List(cl.ctx, nil)
	if err != nil {
		err = fmt.Errorf("listing devices: %w", errorDetail(err))
		return nil, err
	}
	return devices, nil
}

// DeviceShow allows to retrieve a specific device, given its id,
// from Arduino IoT Cloud.
func (cl *client) DeviceShow(id string) (*iotclient.ArduinoDevicev2, error) {
	dev, _, err := cl.api.DevicesV2Api.DevicesV2Show(cl.ctx, id)
	if err != nil {
		err = fmt.Errorf("retrieving device, %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DeviceOTA performs an OTA upload request to Arduino IoT Cloud, passing
// the ID of the device to be updated and the actual file containing the OTA firmware.
func (cl *client) DeviceOTA(id string, file *os.File, expireMins int) error {
	opt := &iotclient.DevicesV2OtaUploadOpts{
		ExpireInMins: optional.NewInt32(int32(expireMins)),
	}
	_, err := cl.api.DevicesV2OtaApi.DevicesV2OtaUpload(cl.ctx, id, file, opt)
	if err != nil {
		err = fmt.Errorf("uploading device ota: %w", errorDetail(err))
		return err
	}
	return nil
}

// CertificateCreate allows to upload a certificate on Arduino IoT Cloud.
// It returns the certificate parameters populated by the cloud.
func (cl *client) CertificateCreate(id, csr string) (*iotclient.ArduinoCompressedv2, error) {
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

// ThingCreate adds a new thing on Arduino IoT Cloud.
func (cl *client) ThingCreate(thing *iotclient.Thing, force bool) (*iotclient.ArduinoThing, error) {
	opt := &iotclient.ThingsV2CreateOpts{Force: optional.NewBool(force)}
	newThing, _, err := cl.api.ThingsV2Api.ThingsV2Create(cl.ctx, *thing, opt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new thing", errorDetail(err))
	}
	return &newThing, nil
}

// ThingUpdate updates a thing on Arduino IoT Cloud.
func (cl *client) ThingUpdate(id string, thing *iotclient.Thing, force bool) error {
	opt := &iotclient.ThingsV2UpdateOpts{Force: optional.NewBool(force)}
	_, _, err := cl.api.ThingsV2Api.ThingsV2Update(cl.ctx, id, *thing, opt)
	if err != nil {
		return fmt.Errorf("%s: %v", "updating thing", errorDetail(err))
	}
	return nil
}

// ThingDelete deletes a thing from Arduino IoT Cloud.
func (cl *client) ThingDelete(id string) error {
	_, err := cl.api.ThingsV2Api.ThingsV2Delete(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting thing: %w", errorDetail(err))
		return err
	}
	return nil
}

// ThingShow allows to retrieve a specific thing, given its id,
// from Arduino IoT Cloud.
func (cl *client) ThingShow(id string) (*iotclient.ArduinoThing, error) {
	thing, _, err := cl.api.ThingsV2Api.ThingsV2Show(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving thing, %w", errorDetail(err))
		return nil, err
	}
	return &thing, nil
}

// ThingList returns a list of things on Arduino IoT Cloud.
func (cl *client) ThingList(ids []string, device *string, props bool) ([]iotclient.ArduinoThing, error) {
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
