// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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

	"github.com/arduino/arduino-cloud-cli/config"
	iotclient "github.com/arduino/iot-client-go/v3"
	"golang.org/x/oauth2"
)

var ErrOtaAlreadyInProgress = fmt.Errorf("ota already in progress")

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

func toStringPointer(s string) *string {
	return &s
}

// DeviceCreate allows to create a new device on Arduino IoT Cloud.
// It returns the newly created device, and an error.
func (cl *Client) DeviceCreate(ctx context.Context, fqbn, name, serial, dType string, cType *string) (*iotclient.ArduinoDevicev2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	payload := iotclient.CreateDevicesV2Payload{
		Fqbn:   toStringPointer(fqbn),
		Name:   toStringPointer(name),
		Serial: toStringPointer(serial),
		Type:   dType,
	}

	if cType != nil {
		payload.ConnectionType = cType
	}

	req := cl.api.DevicesV2API.DevicesV2Create(ctx)
	req = req.CreateDevicesV2Payload(payload)
	dev, _, err := cl.api.DevicesV2API.DevicesV2CreateExecute(req)
	if err != nil {
		err = fmt.Errorf("creating device, %w", errorDetail(err))
		return nil, err
	}
	return dev, nil
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
		Serial:        toStringPointer(serial),
		Type:          devType,
		UserId:        "me",
	}

	req := cl.api.LoraDevicesV1API.LoraDevicesV1Create(ctx)
	req = req.CreateLoraDevicesV1Payload(payload)
	dev, _, err := cl.api.LoraDevicesV1API.LoraDevicesV1CreateExecute(req)
	if err != nil {
		err = fmt.Errorf("creating lora device: %w", errorDetail(err))
		return nil, err
	}
	return dev, nil
}

// DevicePassSet sets the device password to the one suggested by Arduino IoT Cloud.
// Returns the set password.
func (cl *Client) DevicePassSet(ctx context.Context, id string) (*iotclient.ArduinoDevicev2Pass, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	// Fetch suggested password
	req := cl.api.DevicesV2PassAPI.DevicesV2PassGet(ctx, id)
	req = req.SuggestedPassword(true)
	pass, _, err := cl.api.DevicesV2PassAPI.DevicesV2PassGetExecute(req)
	if err != nil {
		err = fmt.Errorf("fetching device suggested password: %w", errorDetail(err))
		return nil, err
	}

	// Set password to the suggested one
	reqSet := cl.api.DevicesV2PassAPI.DevicesV2PassSet(ctx, id)
	reqSet = reqSet.Devicev2Pass(iotclient.Devicev2Pass{Password: pass.SuggestedPassword})
	pass, _, err = cl.api.DevicesV2PassAPI.DevicesV2PassSetExecute(reqSet)
	if err != nil {
		err = fmt.Errorf("setting device password: %w", errorDetail(err))
		return nil, err
	}
	return pass, nil
}

// DeviceDelete deletes the device corresponding to the passed ID
// from Arduino IoT Cloud.
func (cl *Client) DeviceDelete(ctx context.Context, id string) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	req := cl.api.DevicesV2API.DevicesV2Delete(ctx, id)
	_, err = cl.api.DevicesV2API.DevicesV2DeleteExecute(req)
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

	req := cl.api.DevicesV2API.DevicesV2List(ctx)
	if tags != nil {
		t := make([]string, 0, len(tags))
		for key, val := range tags {
			// Use the 'key:value' format required from the backend
			t = append(t, key+":"+val)
		}
		req = req.Tags(t)
	}
	devices, _, err := cl.api.DevicesV2API.DevicesV2ListExecute(req)
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

	req := cl.api.DevicesV2API.DevicesV2Show(ctx, id)
	dev, _, err := cl.api.DevicesV2API.DevicesV2ShowExecute(req)
	if err != nil {
		err = fmt.Errorf("retrieving device, %w", errorDetail(err))
		return nil, err
	}
	return dev, nil
}

// DeviceNetworkCredentials allows to retrieve a specific device network credentials configuration options
func (cl *Client) DeviceNetworkCredentials(ctx context.Context, deviceType, connection string) ([]iotclient.ArduinoCredentialsv1, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	req := cl.api.NetworkCredentialsV1API.NetworkCredentialsV1Show(ctx, deviceType)
	req = req.Connection(connection)
	dev, _, err := cl.api.NetworkCredentialsV1API.NetworkCredentialsV1ShowExecute(req)
	if err != nil {
		err = fmt.Errorf("retrieving device network configuration, %w", errorDetail(err))
		return nil, err
	}
	return dev, nil
}

// DeviceOTA performs an OTA upload request to Arduino IoT Cloud, passing
// the ID of the device to be updated and the actual file containing the OTA firmware.
func (cl *Client) DeviceOTA(ctx context.Context, id string, file *os.File, expireMins int) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	req := cl.api.DevicesV2OtaAPI.DevicesV2OtaUpload(ctx, id)
	req = req.ExpireInMins(int32(expireMins))
	req = req.Async(true)
	req = req.OtaFile(file)
	_, resp, err := cl.api.DevicesV2OtaAPI.DevicesV2OtaUploadExecute(req)
	if err != nil {
		// 409 (Conflict) is the status code for an already existing OTA in progress for the same device. Handling it in a different way.
		if resp != nil && resp.StatusCode == 409 {
			return ErrOtaAlreadyInProgress
		}
		return fmt.Errorf("uploading device ota: %w", errorDetail(err))
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
		req := cl.api.DevicesV2TagsAPI.DevicesV2TagsUpsert(ctx, id)
		req = req.Tag(iotclient.Tag{Key: key, Value: val})
		_, err := cl.api.DevicesV2TagsAPI.DevicesV2TagsUpsertExecute(req)
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
		req := cl.api.DevicesV2TagsAPI.DevicesV2TagsDelete(ctx, id, key)
		_, err := cl.api.DevicesV2TagsAPI.DevicesV2TagsDeleteExecute(req)
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

	req := cl.api.LoraFreqPlanV1API.LoraFreqPlanV1List(ctx)
	freqs, _, err := cl.api.LoraFreqPlanV1API.LoraFreqPlanV1ListExecute(req)
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
		Ca:      toStringPointer("Arduino_v2"),
		Csr:     csr,
		Enabled: true,
	}

	req := cl.api.DevicesV2CertsAPI.DevicesV2CertsCreate(ctx, id)
	req = req.CreateDevicesV2CertsPayload(cert)
	newCert, _, err := cl.api.DevicesV2CertsAPI.DevicesV2CertsCreateExecute(req)
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

	req := cl.api.ThingsV2API.ThingsV2Create(ctx)
	req = req.ThingCreate(*thing)
	req = req.Force(force)
	newThing, _, err := cl.api.ThingsV2API.ThingsV2CreateExecute(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new thing", errorDetail(err))
	}
	return newThing, nil
}

// ThingUpdate updates a thing on Arduino IoT Cloud.
func (cl *Client) ThingUpdate(ctx context.Context, id string, thing *iotclient.ThingUpdate, force bool) error {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return err
	}

	req := cl.api.ThingsV2API.ThingsV2Update(ctx, id)
	req = req.Force(force)
	req = req.ThingUpdate(*thing)
	_, _, err = cl.api.ThingsV2API.ThingsV2UpdateExecute(req)
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

	req := cl.api.ThingsV2API.ThingsV2Delete(ctx, id)
	_, err = cl.api.ThingsV2API.ThingsV2DeleteExecute(req)
	if err != nil {
		err = fmt.Errorf("deleting thing: %w", errorDetail(err))
		return err
	}
	return nil
}

// ThingShow allows to retrieve a specific thing, given its id from Arduino IoT Cloud.
func (cl *Client) ThingShow(ctx context.Context, id string) (*iotclient.ArduinoThing, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	req := cl.api.ThingsV2API.ThingsV2Show(ctx, id)
	thing, _, err := cl.api.ThingsV2API.ThingsV2ShowExecute(req)
	if err != nil {
		return nil, fmt.Errorf("retrieving thing, %w", errorDetail(err))
	}
	return thing, nil
}

// ThingClone allows to clone a specific thing, given its id from Arduino IoT Cloud.
func (cl *Client) ThingClone(ctx context.Context, id, newName string) (*iotclient.ArduinoThing, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	req := cl.api.ThingsV2API.ThingsV2Clone(ctx, id)
	includeTags := true
	req = req.ThingClone(iotclient.ThingClone{Name: newName, IncludeTags: &includeTags})
	thing, _, err := cl.api.ThingsV2API.ThingsV2CloneExecute(req)
	if err != nil {
		return nil, fmt.Errorf("cloning thing thing, %w", errorDetail(err))
	}
	return thing, nil
}

// ThingList returns a list of things on Arduino IoT Cloud.
func (cl *Client) ThingList(ctx context.Context, ids []string, device *string, props bool, tags map[string]string) ([]iotclient.ArduinoThing, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	req := cl.api.ThingsV2API.ThingsV2List(ctx)
	req = req.ShowProperties(props)
	if ids != nil {
		req = req.Ids(ids)
	}

	if device != nil {
		req = req.DeviceId(*device)
	}

	if tags != nil {
		t := make([]string, 0, len(tags))
		for key, val := range tags {
			// Use the 'key:value' format required from the backend
			t = append(t, key+":"+val)
		}
		req = req.Tags(t)
	}
	things, _, err := cl.api.ThingsV2API.ThingsV2ListExecute(req)
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
		req := cl.api.ThingsV2TagsAPI.ThingsV2TagsUpsert(ctx, id)
		req = req.Tag(iotclient.Tag{Key: key, Value: val})
		_, err := cl.api.ThingsV2TagsAPI.ThingsV2TagsUpsertExecute(req)
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
		req := cl.api.ThingsV2TagsAPI.ThingsV2TagsDelete(ctx, id, key)
		_, err := cl.api.ThingsV2TagsAPI.ThingsV2TagsDeleteExecute(req)
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

	req := cl.api.DashboardsV2API.DashboardsV2Create(ctx)
	req = req.Dashboardv2(*dashboard)
	newDashboard, _, err := cl.api.DashboardsV2API.DashboardsV2CreateExecute(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "adding new dashboard", errorDetail(err))
	}
	return newDashboard, nil
}

// DashboardShow allows to retrieve a specific dashboard, given its id,
// from Arduino IoT Cloud.
func (cl *Client) DashboardShow(ctx context.Context, id string) (*iotclient.ArduinoDashboardv2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	req := cl.api.DashboardsV2API.DashboardsV2Show(ctx, id)
	dashboard, _, err := cl.api.DashboardsV2API.DashboardsV2ShowExecute(req)
	if err != nil {
		err = fmt.Errorf("retrieving dashboard, %w", errorDetail(err))
		return nil, err
	}
	return dashboard, nil
}

// DashboardList returns a list of dashboards on Arduino IoT Cloud.
func (cl *Client) DashboardList(ctx context.Context) ([]iotclient.ArduinoDashboardv2, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	req := cl.api.DashboardsV2API.DashboardsV2List(ctx)
	dashboards, _, err := cl.api.DashboardsV2API.DashboardsV2ListExecute(req)
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

	req := cl.api.DashboardsV2API.DashboardsV2Delete(ctx, id)
	_, err = cl.api.DashboardsV2API.DashboardsV2DeleteExecute(req)
	if err != nil {
		err = fmt.Errorf("deleting dashboard: %w", errorDetail(err))
		return err
	}
	return nil
}

// TemplateApply apply a given template, creating associated resources like things and dashboards.
func (cl *Client) TemplateApply(ctx context.Context, id, thingId, prefix, deviceId string, credentials map[string]string) (*iotclient.ArduinoTemplate, error) {
	ctx, err := ctxWithToken(ctx, cl.token)
	if err != nil {
		return nil, err
	}

	thingOption := make(map[string]any)
	thingOption["device_id"] = deviceId
	thingOption["secrets"] = credentials

	req := cl.api.TemplatesAPI.TemplatesApply(ctx)
	req = req.Template(iotclient.Template{
		PrefixName:       toStringPointer(prefix),
		CustomTemplateId: toStringPointer(id),
		ThingsOptions: map[string]interface{}{
			thingId: thingOption,
		},
	})
	dev, _, err := cl.api.TemplatesAPI.TemplatesApplyExecute(req)
	if err != nil {
		err = fmt.Errorf("retrieving device, %w", errorDetail(err))
		return nil, err
	}
	return dev, nil
}

func (cl *Client) setup(client, secret, organizationId string) error {
	baseURL := GetArduinoAPIBaseURL()

	// Configure a token source given the user's credentials.
	cl.token = NewUserTokenSource(client, secret, baseURL, organizationId)

	config := iotclient.NewConfiguration()
	if organizationId != "" {
		config.AddDefaultHeader("X-Organization", organizationId)
	}
	config.Servers = iotclient.ServerConfigurations{
		{
			URL:         baseURL,
			Description: "IoT API endpoint",
		},
	}
	cl.api = iotclient.NewAPIClient(config)

	return nil
}
