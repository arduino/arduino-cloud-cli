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
	"github.com/arduino/arduino-cloud-cli/config"
	iotclient "github.com/arduino/iot-client-go"
)

// Client can perform actions on Arduino IoT Cloud.
type Client struct {
	ctx context.Context
	api *iotclient.APIClient
}

// NewClient returns a new client implementing the Client interface.
// It needs client Credentials for cloud authentication.
func NewClient(cred *config.Credentials) (*Client, error) {
	cl := &Client{}
	err := cl.setup(cred.Client, cred.Secret, cred.Organization)
	if err != nil {
		err = fmt.Errorf("instantiate new iot client: %w", err)
		return nil, err
	}
	return cl, nil
}

// DeviceCreate allows to create a new device on Arduino IoT Cloud.
// It returns the newly created device, and an error.
func (cl *Client) DeviceCreate(fqbn, name, serial, dType string) (*iotclient.ArduinoDevicev2, error) {
	payload := iotclient.CreateDevicesV2Payload{
		Fqbn:   fqbn,
		Name:   name,
		Serial: serial,
		Type:   dType,
	}
	dev, _, err := cl.api.DevicesV2Api.DevicesV2Create(cl.ctx, payload, nil)
	if err != nil {
		err = fmt.Errorf("creating device, %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DeviceLoraCreate allows to create a new LoRa device on Arduino IoT Cloud.
// It returns the LoRa information about the newly created device, and an error.
func (cl *Client) DeviceLoraCreate(name, serial, devType, eui, freq string) (*iotclient.ArduinoLoradevicev1, error) {
	payload := iotclient.CreateLoraDevicesV1Payload{
		App:           "defaultApp",
		Eui:           eui,
		FrequencyPlan: freq,
		Name:          name,
		Serial:        serial,
		Type:          devType,
		UserId:        "me",
	}
	dev, _, err := cl.api.LoraDevicesV1Api.LoraDevicesV1Create(cl.ctx, payload)
	if err != nil {
		err = fmt.Errorf("creating lora device: %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DevicePassSet sets the device password to the one suggested by Arduino IoT Cloud.
// Returns the set password.
func (cl *Client) DevicePassSet(id string) (*iotclient.ArduinoDevicev2Pass, error) {
	// Fetch suggested password
	opts := &iotclient.DevicesV2PassGetOpts{SuggestedPassword: optional.NewBool(true)}
	pass, _, err := cl.api.DevicesV2PassApi.DevicesV2PassGet(cl.ctx, id, opts)
	if err != nil {
		err = fmt.Errorf("fetching device suggested password: %w", errorDetail(err))
		return nil, err
	}
	// Set password to the suggested one
	p := iotclient.Devicev2Pass{Password: pass.SuggestedPassword}
	pass, _, err = cl.api.DevicesV2PassApi.DevicesV2PassSet(cl.ctx, id, p)
	if err != nil {
		err = fmt.Errorf("setting device password: %w", errorDetail(err))
		return nil, err
	}
	return &pass, nil
}

// DeviceDelete deletes the device corresponding to the passed ID
// from Arduino IoT Cloud.
func (cl *Client) DeviceDelete(id string) error {
	_, err := cl.api.DevicesV2Api.DevicesV2Delete(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting device: %w", errorDetail(err))
		return err
	}
	return nil
}

// DeviceList retrieves and returns a list of all Arduino IoT Cloud devices
// belonging to the user performing the request.
func (cl *Client) DeviceList(tags map[string]string) ([]iotclient.ArduinoDevicev2, error) {
	opts := &iotclient.DevicesV2ListOpts{}
	if tags != nil {
		t := make([]string, 0, len(tags))
		for key, val := range tags {
			// Use the 'key:value' format required from the backend
			t = append(t, key+":"+val)
		}
		opts.Tags = optional.NewInterface(t)
	}

	devices, _, err := cl.api.DevicesV2Api.DevicesV2List(cl.ctx, opts)
	if err != nil {
		err = fmt.Errorf("listing devices: %w", errorDetail(err))
		return nil, err
	}
	return devices, nil
}

// DeviceShow allows to retrieve a specific device, given its id,
// from Arduino IoT Cloud.
func (cl *Client) DeviceShow(id string) (*iotclient.ArduinoDevicev2, error) {
	dev, _, err := cl.api.DevicesV2Api.DevicesV2Show(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving device, %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DeviceOTA performs an OTA upload request to Arduino IoT Cloud, passing
// the ID of the device to be updated and the actual file containing the OTA firmware.
func (cl *Client) DeviceOTA(id string, file *os.File, expireMins int) error {
	opt := &iotclient.DevicesV2OtaUploadOpts{
		ExpireInMins: optional.NewInt32(int32(expireMins)),
		Async:        optional.NewBool(true),
	}
	_, err := cl.api.DevicesV2OtaApi.DevicesV2OtaUpload(cl.ctx, id, file, opt)
	if err != nil {
		err = fmt.Errorf("uploading device ota: %w", errorDetail(err))
		return err
	}
	return nil
}

// DeviceTagsCreate allows to create or overwrite tags on a device of Arduino IoT Cloud.
func (cl *Client) DeviceTagsCreate(id string, tags map[string]string) error {
	for key, val := range tags {
		t := iotclient.Tag{Key: key, Value: val}
		_, err := cl.api.DevicesV2TagsApi.DevicesV2TagsUpsert(cl.ctx, id, t)
		if err != nil {
			err = fmt.Errorf("cannot create tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// DeviceTagsDelete deletes the tags of a device of Arduino IoT Cloud,
// given the device id and the keys of the tags.
func (cl *Client) DeviceTagsDelete(id string, keys []string) error {
	for _, key := range keys {
		_, err := cl.api.DevicesV2TagsApi.DevicesV2TagsDelete(cl.ctx, id, key)
		if err != nil {
			err = fmt.Errorf("cannot delete tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// LoraFrequencyPlansList retrieves and returns the list of all supported
// LoRa frequency plans.
func (cl *Client) LoraFrequencyPlansList() ([]iotclient.ArduinoLorafreqplanv1, error) {
	freqs, _, err := cl.api.LoraFreqPlanV1Api.LoraFreqPlanV1List(cl.ctx)
	if err != nil {
		err = fmt.Errorf("listing lora frequency plans: %w", errorDetail(err))
		return nil, err
	}
	return freqs.FrequencyPlans, nil
}

// CertificateCreate allows to upload a certificate on Arduino IoT Cloud.
// It returns the certificate parameters populated by the cloud.
func (cl *Client) CertificateCreate(id, csr string) (*iotclient.ArduinoDevicev2Cert, error) {
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

	return &newCert, nil
}

// ThingCreate adds a new thing on Arduino IoT Cloud.
func (cl *Client) ThingCreate(thing *iotclient.ThingCreate, force bool) (*iotclient.ArduinoThing, error) {
	opt := &iotclient.ThingsV2CreateOpts{Force: optional.NewBool(force)}
	newThing, _, err := cl.api.ThingsV2Api.ThingsV2Create(cl.ctx, *thing, opt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new thing", errorDetail(err))
	}
	return &newThing, nil
}

// ThingUpdate updates a thing on Arduino IoT Cloud.
func (cl *Client) ThingUpdate(id string, thing *iotclient.ThingUpdate, force bool) error {
	opt := &iotclient.ThingsV2UpdateOpts{Force: optional.NewBool(force)}
	_, _, err := cl.api.ThingsV2Api.ThingsV2Update(cl.ctx, id, *thing, opt)
	if err != nil {
		return fmt.Errorf("%s: %v", "updating thing", errorDetail(err))
	}
	return nil
}

// ThingDelete deletes a thing from Arduino IoT Cloud.
func (cl *Client) ThingDelete(id string) error {
	_, err := cl.api.ThingsV2Api.ThingsV2Delete(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting thing: %w", errorDetail(err))
		return err
	}
	return nil
}

// ThingShow allows to retrieve a specific thing, given its id,
// from Arduino IoT Cloud.
func (cl *Client) ThingShow(id string) (*iotclient.ArduinoThing, error) {
	thing, _, err := cl.api.ThingsV2Api.ThingsV2Show(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving thing, %w", errorDetail(err))
		return nil, err
	}
	return &thing, nil
}

// ThingList returns a list of things on Arduino IoT Cloud.
func (cl *Client) ThingList(ids []string, device *string, props bool, tags map[string]string) ([]iotclient.ArduinoThing, error) {
	opts := &iotclient.ThingsV2ListOpts{}
	opts.ShowProperties = optional.NewBool(props)

	if ids != nil {
		opts.Ids = optional.NewInterface(ids)
	}

	if device != nil {
		opts.DeviceId = optional.NewString(*device)
	}

	if tags != nil {
		t := make([]string, 0, len(tags))
		for key, val := range tags {
			// Use the 'key:value' format required from the backend
			t = append(t, key+":"+val)
		}
		opts.Tags = optional.NewInterface(t)
	}

	things, _, err := cl.api.ThingsV2Api.ThingsV2List(cl.ctx, opts)
	if err != nil {
		err = fmt.Errorf("retrieving things, %w", errorDetail(err))
		return nil, err
	}
	return things, nil
}

// ThingTagsCreate allows to create or overwrite tags on a thing of Arduino IoT Cloud.
func (cl *Client) ThingTagsCreate(id string, tags map[string]string) error {
	for key, val := range tags {
		t := iotclient.Tag{Key: key, Value: val}
		_, err := cl.api.ThingsV2TagsApi.ThingsV2TagsUpsert(cl.ctx, id, t)
		if err != nil {
			err = fmt.Errorf("cannot create tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// ThingTagsDelete deletes the tags of a thing of Arduino IoT Cloud,
// given the thing id and the keys of the tags.
func (cl *Client) ThingTagsDelete(id string, keys []string) error {
	for _, key := range keys {
		_, err := cl.api.ThingsV2TagsApi.ThingsV2TagsDelete(cl.ctx, id, key)
		if err != nil {
			err = fmt.Errorf("cannot delete tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// DashboardCreate adds a new dashboard on Arduino IoT Cloud.
func (cl *Client) DashboardCreate(dashboard *iotclient.Dashboardv2) (*iotclient.ArduinoDashboardv2, error) {
	newDashboard, _, err := cl.api.DashboardsV2Api.DashboardsV2Create(cl.ctx, *dashboard, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new dashboard", errorDetail(err))
	}
	return &newDashboard, nil
}

// DashboardShow allows to retrieve a specific dashboard, given its id,
// from Arduino IoT Cloud.
func (cl *Client) DashboardShow(id string) (*iotclient.ArduinoDashboardv2, error) {
	dashboard, _, err := cl.api.DashboardsV2Api.DashboardsV2Show(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving dashboard, %w", errorDetail(err))
		return nil, err
	}
	return &dashboard, nil
}

// DashboardList returns a list of dashboards on Arduino IoT Cloud.
func (cl *Client) DashboardList() ([]iotclient.ArduinoDashboardv2, error) {
	dashboards, _, err := cl.api.DashboardsV2Api.DashboardsV2List(cl.ctx, nil)
	if err != nil {
		err = fmt.Errorf("listing dashboards: %w", errorDetail(err))
		return nil, err
	}
	return dashboards, nil
}

// DashboardDelete deletes a dashboard from Arduino IoT Cloud.
func (cl *Client) DashboardDelete(id string) error {
	_, err := cl.api.DashboardsV2Api.DashboardsV2Delete(cl.ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting dashboard: %w", errorDetail(err))
		return err
	}
	return nil
}

func (cl *Client) setup(client, secret, organization string) error {
	baseURL := "https://api2.arduino.cc"
	if url := os.Getenv("IOT_API_URL"); url != "" {
		baseURL = url
	}

	// Get the access token in exchange of client_id and client_secret
	tok, err := token(client, secret, baseURL)
	if err != nil {
		err = fmt.Errorf("cannot retrieve token given client and secret: %w", err)
		return err
	}

	// We use the token to create a context that will be passed to any API call
	cl.ctx = context.WithValue(context.Background(), iotclient.ContextAccessToken, tok.AccessToken)

	config := iotclient.NewConfiguration()
	if organization != "" {
		config.DefaultHeader = map[string]string{"X-Organization": organization}
	}
	config.BasePath = baseURL + "/iot"
	cl.api = iotclient.NewAPIClient(config)

	return nil
}
