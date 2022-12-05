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
	"golang.org/x/oauth2"
)

// Client can perform actions on Arduino IoT Cloud.
type Client struct {
	api   *iotclient.APIClient
	token oauth2.TokenSource
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
func (cl *Client) DeviceCreate(ctx context.Context, fqbn, name, serial, dType string) (*iotclient.ArduinoDevicev2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	payload := iotclient.CreateDevicesV2Payload{
		Fqbn:   fqbn,
		Name:   name,
		Serial: serial,
		Type:   dType,
	}
	dev, _, err := cl.api.DevicesV2Api.DevicesV2Create(ctx, payload, nil)
	if err != nil {
		err = fmt.Errorf("creating device, %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DeviceLoraCreate allows to create a new LoRa device on Arduino IoT Cloud.
// It returns the LoRa information about the newly created device, and an error.
func (cl *Client) DeviceLoraCreate(ctx context.Context, name, serial, devType, eui, freq string) (*iotclient.ArduinoLoradevicev1, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	payload := iotclient.CreateLoraDevicesV1Payload{
		App:           "defaultApp",
		Eui:           eui,
		FrequencyPlan: freq,
		Name:          name,
		Serial:        serial,
		Type:          devType,
		UserId:        "me",
	}
	dev, _, err := cl.api.LoraDevicesV1Api.LoraDevicesV1Create(ctx, payload)
	if err != nil {
		err = fmt.Errorf("creating lora device: %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DevicePassSet sets the device password to the one suggested by Arduino IoT Cloud.
// Returns the set password.
func (cl *Client) DevicePassSet(ctx context.Context, id string) (*iotclient.ArduinoDevicev2Pass, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	// Fetch suggested password
	opts := &iotclient.DevicesV2PassGetOpts{SuggestedPassword: optional.NewBool(true)}
	pass, _, err := cl.api.DevicesV2PassApi.DevicesV2PassGet(ctx, id, opts)
	if err != nil {
		err = fmt.Errorf("fetching device suggested password: %w", errorDetail(err))
		return nil, err
	}

	// Set password to the suggested one
	p := iotclient.Devicev2Pass{Password: pass.SuggestedPassword}
	pass, _, err = cl.api.DevicesV2PassApi.DevicesV2PassSet(ctx, id, p)
	if err != nil {
		err = fmt.Errorf("setting device password: %w", errorDetail(err))
		return nil, err
	}
	return &pass, nil
}

// DeviceDelete deletes the device corresponding to the passed ID
// from Arduino IoT Cloud.
func (cl *Client) DeviceDelete(ctx context.Context, id string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	_, err = cl.api.DevicesV2Api.DevicesV2Delete(ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting device: %w", errorDetail(err))
		return err
	}
	return nil
}

// DeviceList retrieves and returns a list of all Arduino IoT Cloud devices
// belonging to the user performing the request.
func (cl *Client) DeviceList(ctx context.Context, tags map[string]string) ([]iotclient.ArduinoDevicev2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	opts := &iotclient.DevicesV2ListOpts{}
	if tags != nil {
		t := make([]string, 0, len(tags))
		for key, val := range tags {
			// Use the 'key:value' format required from the backend
			t = append(t, key+":"+val)
		}
		opts.Tags = optional.NewInterface(t)
	}

	devices, _, err := cl.api.DevicesV2Api.DevicesV2List(ctx, opts)
	if err != nil {
		err = fmt.Errorf("listing devices: %w", errorDetail(err))
		return nil, err
	}
	return devices, nil
}

// DeviceShow allows to retrieve a specific device, given its id,
// from Arduino IoT Cloud.
func (cl *Client) DeviceShow(ctx context.Context, id string) (*iotclient.ArduinoDevicev2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	dev, _, err := cl.api.DevicesV2Api.DevicesV2Show(ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving device, %w", errorDetail(err))
		return nil, err
	}
	return &dev, nil
}

// DeviceOTA performs an OTA upload request to Arduino IoT Cloud, passing
// the ID of the device to be updated and the actual file containing the OTA firmware.
func (cl *Client) DeviceOTA(ctx context.Context, id string, file *os.File, expireMins int) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	opt := &iotclient.DevicesV2OtaUploadOpts{
		ExpireInMins: optional.NewInt32(int32(expireMins)),
		Async:        optional.NewBool(true),
	}
	_, err = cl.api.DevicesV2OtaApi.DevicesV2OtaUpload(ctx, id, file, opt)
	if err != nil {
		err = fmt.Errorf("uploading device ota: %w", errorDetail(err))
		return err
	}
	return nil
}

// DeviceTagsCreate allows to create or overwrite tags on a device of Arduino IoT Cloud.
func (cl *Client) DeviceTagsCreate(ctx context.Context, id string, tags map[string]string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	for key, val := range tags {
		t := iotclient.Tag{Key: key, Value: val}
		_, err := cl.api.DevicesV2TagsApi.DevicesV2TagsUpsert(ctx, id, t)
		if err != nil {
			err = fmt.Errorf("cannot create tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// DeviceTagsDelete deletes the tags of a device of Arduino IoT Cloud,
// given the device id and the keys of the tags.
func (cl *Client) DeviceTagsDelete(ctx context.Context, id string, keys []string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err := cl.api.DevicesV2TagsApi.DevicesV2TagsDelete(ctx, id, key)
		if err != nil {
			err = fmt.Errorf("cannot delete tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// LoraFrequencyPlansList retrieves and returns the list of all supported
// LoRa frequency plans.
func (cl *Client) LoraFrequencyPlansList(ctx context.Context) ([]iotclient.ArduinoLorafreqplanv1, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	freqs, _, err := cl.api.LoraFreqPlanV1Api.LoraFreqPlanV1List(ctx)
	if err != nil {
		err = fmt.Errorf("listing lora frequency plans: %w", errorDetail(err))
		return nil, err
	}
	return freqs.FrequencyPlans, nil
}

// CertificateCreate allows to upload a certificate on Arduino IoT Cloud.
// It returns the certificate parameters populated by the cloud.
func (cl *Client) CertificateCreate(ctx context.Context, id, csr string) (*iotclient.ArduinoCompressedv2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	cert := iotclient.CreateDevicesV2CertsPayload{
		Ca:      "Arduino",
		Csr:     csr,
		Enabled: true,
	}

	newCert, _, err := cl.api.DevicesV2CertsApi.DevicesV2CertsCreate(ctx, id, cert)
	if err != nil {
		err = fmt.Errorf("creating certificate, %w", errorDetail(err))
		return nil, err
	}

	return &newCert.Compressed, nil
}

// ThingCreate adds a new thing on Arduino IoT Cloud.
func (cl *Client) ThingCreate(ctx context.Context, thing *iotclient.ThingCreate, force bool) (*iotclient.ArduinoThing, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	opt := &iotclient.ThingsV2CreateOpts{Force: optional.NewBool(force)}
	newThing, _, err := cl.api.ThingsV2Api.ThingsV2Create(ctx, *thing, opt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new thing", errorDetail(err))
	}
	return &newThing, nil
}

// ThingUpdate updates a thing on Arduino IoT Cloud.
func (cl *Client) ThingUpdate(ctx context.Context, id string, thing *iotclient.ThingUpdate, force bool) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	opt := &iotclient.ThingsV2UpdateOpts{Force: optional.NewBool(force)}
	_, _, err = cl.api.ThingsV2Api.ThingsV2Update(ctx, id, *thing, opt)
	if err != nil {
		return fmt.Errorf("%s: %v", "updating thing", errorDetail(err))
	}
	return nil
}

// ThingDelete deletes a thing from Arduino IoT Cloud.
func (cl *Client) ThingDelete(ctx context.Context, id string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	_, err = cl.api.ThingsV2Api.ThingsV2Delete(ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("deleting thing: %w", errorDetail(err))
		return err
	}
	return nil
}

// ThingShow allows to retrieve a specific thing, given its id,
// from Arduino IoT Cloud.
func (cl *Client) ThingShow(ctx context.Context, id string) (*iotclient.ArduinoThing, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	thing, _, err := cl.api.ThingsV2Api.ThingsV2Show(ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving thing, %w", errorDetail(err))
		return nil, err
	}
	return &thing, nil
}

// ThingList returns a list of things on Arduino IoT Cloud.
func (cl *Client) ThingList(ctx context.Context, ids []string, device *string, props bool, tags map[string]string) ([]iotclient.ArduinoThing, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

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

	things, _, err := cl.api.ThingsV2Api.ThingsV2List(ctx, opts)
	if err != nil {
		err = fmt.Errorf("retrieving things, %w", errorDetail(err))
		return nil, err
	}
	return things, nil
}

// ThingTagsCreate allows to create or overwrite tags on a thing of Arduino IoT Cloud.
func (cl *Client) ThingTagsCreate(ctx context.Context, id string, tags map[string]string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	for key, val := range tags {
		t := iotclient.Tag{Key: key, Value: val}
		_, err := cl.api.ThingsV2TagsApi.ThingsV2TagsUpsert(ctx, id, t)
		if err != nil {
			err = fmt.Errorf("cannot create tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// ThingTagsDelete deletes the tags of a thing of Arduino IoT Cloud,
// given the thing id and the keys of the tags.
func (cl *Client) ThingTagsDelete(ctx context.Context, id string, keys []string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err := cl.api.ThingsV2TagsApi.ThingsV2TagsDelete(ctx, id, key)
		if err != nil {
			err = fmt.Errorf("cannot delete tag %s: %w", key, errorDetail(err))
			return err
		}
	}
	return nil
}

// DashboardCreate adds a new dashboard on Arduino IoT Cloud.
func (cl *Client) DashboardCreate(ctx context.Context, dashboard *iotclient.Dashboardv2) (*iotclient.ArduinoDashboardv2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	newDashboard, _, err := cl.api.DashboardsV2Api.DashboardsV2Create(ctx, *dashboard, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new dashboard", errorDetail(err))
	}
	return &newDashboard, nil
}

// DashboardShow allows to retrieve a specific dashboard, given its id,
// from Arduino IoT Cloud.
func (cl *Client) DashboardShow(ctx context.Context, id string) (*iotclient.ArduinoDashboardv2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	dashboard, _, err := cl.api.DashboardsV2Api.DashboardsV2Show(ctx, id, nil)
	if err != nil {
		err = fmt.Errorf("retrieving dashboard, %w", errorDetail(err))
		return nil, err
	}
	return &dashboard, nil
}

// DashboardList returns a list of dashboards on Arduino IoT Cloud.
func (cl *Client) DashboardList(ctx context.Context) ([]iotclient.ArduinoDashboardv2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	dashboards, _, err := cl.api.DashboardsV2Api.DashboardsV2List(ctx, nil)
	if err != nil {
		err = fmt.Errorf("listing dashboards: %w", errorDetail(err))
		return nil, err
	}
	return dashboards, nil
}

// DashboardDelete deletes a dashboard from Arduino IoT Cloud.
func (cl *Client) DashboardDelete(ctx context.Context, id string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	_, err = cl.api.DashboardsV2Api.DashboardsV2Delete(ctx, id, nil)
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

	// Configure a token source given the user's credentials.
	cl.token = token(client, secret, baseURL)

	config := iotclient.NewConfiguration()
	if organization != "" {
		config.DefaultHeader = map[string]string{"X-Organization": organization}
	}
	config.BasePath = baseURL + "/iot"
	cl.api = iotclient.NewAPIClient(config)

	return nil
}
