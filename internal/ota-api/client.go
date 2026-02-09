// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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

package otaapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"golang.org/x/oauth2"
)

const (
	OrderDesc = "desc"
	OrderAsc  = "asc"
)

var ErrAlreadyInProgress = fmt.Errorf("already in progress")
var ErrAlreadyCancelled = fmt.Errorf("already cancelled")

type OtaApiClient struct {
	client       *http.Client
	host         string
	src          oauth2.TokenSource
	organization string
}

func NewClient(credentials *config.Credentials) *OtaApiClient {
	host := iot.GetArduinoAPIBaseURL()
	tokenSource := iot.NewUserTokenSource(credentials.Client, credentials.Secret, host, credentials.Organization)
	return &OtaApiClient{
		client:       &http.Client{},
		src:          tokenSource,
		host:         host,
		organization: credentials.Organization,
	}
}

func (c *OtaApiClient) performGetRequest(endpoint, token string) (*http.Response, error) {
	return c.performRequest(endpoint, "GET", token)
}

func (c *OtaApiClient) performRequest(endpoint, method, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	if c.organization != "" {
		req.Header.Add("X-Organization", c.organization)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *OtaApiClient) GetOtaStatusByOtaID(otaid string, limit int, order string) (*OtaStatusResponse, error) {

	if otaid == "" {
		return nil, fmt.Errorf("invalid ota-id: empty")
	}

	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	endpoint := c.host + "/ota/v1/ota/" + otaid
	res, err := c.performGetRequest(endpoint, userRequestToken.AccessToken)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyb, err := io.ReadAll(res.Body)

	if res.StatusCode == 200 {
		var otaResponse OtaStatusResponse
		if err == nil && bodyb != nil {
			err = json.Unmarshal(bodyb, &otaResponse)
			if err != nil {
				return nil, err
			}
		}

		if len(otaResponse.States) > 0 {
			// Sort output by StartedAt
			sort.Slice(otaResponse.States, func(i, j int) bool {
				t1, err := time.Parse(time.RFC3339, otaResponse.States[i].Timestamp)
				if err != nil {
					return false
				}
				t2, err := time.Parse(time.RFC3339, otaResponse.States[j].Timestamp)
				if err != nil {
					return false
				}
				if order == "asc" {
					return t1.Before(t2)
				}
				return t1.After(t2)
			})
			if limit > 0 && len(otaResponse.States) > limit {
				otaResponse.States = otaResponse.States[:limit]
			}
		}

		return &otaResponse, nil
	} else if res.StatusCode == 404 || res.StatusCode == 400 {
		return nil, fmt.Errorf("ota-id %s not found", otaid)
	}

	return nil, err
}

func (c *OtaApiClient) GetOtaStatusByOtaIDs(otaids string) (*OtaStatusList, error) {

	ids := strings.Split(otaids, ",")
	if len(ids) == 0 {
		return nil, fmt.Errorf("invalid ota-ids: empty")
	}

	returnStatus := OtaStatusList{}
	for _, id := range ids {
		if id != "" {
			resp, err := c.GetOtaStatusByOtaID(id, 1, OrderDesc)
			if err != nil {
				return nil, err
			}
			returnStatus.Ota = append(returnStatus.Ota, resp.Ota)
		}

	}

	return &returnStatus, nil
}

func (c *OtaApiClient) GetOtaLastStatusByDeviceID(deviceID string) (*OtaStatusList, error) {
	return c.GetOtaStatusByDeviceID(deviceID, 1, OrderDesc)
}

func (c *OtaApiClient) GetOtaStatusByDeviceID(deviceID string, limit int, order string) (*OtaStatusList, error) {

	if deviceID == "" {
		return nil, fmt.Errorf("invalid device-id: empty")
	}

	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	endpoint := c.host + "/ota/v1/ota?device_id=" + deviceID
	if limit > 0 {
		endpoint += "&limit=" + fmt.Sprintf("%d", limit)
	}
	if order != "" && (order == "asc" || order == "desc") {
		endpoint += "&order=" + order
	}
	res, err := c.performGetRequest(endpoint, userRequestToken.AccessToken)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyb, err := io.ReadAll(res.Body)

	if res.StatusCode == 200 {
		var otaResponse OtaStatusList
		if err == nil && bodyb != nil {
			err = json.Unmarshal(bodyb, &otaResponse)
			if err != nil {
				return nil, err
			}
		}
		return &otaResponse, nil
	} else if res.StatusCode == 404 || res.StatusCode == 400 {
		return nil, fmt.Errorf("device-id %s not found", deviceID)
	} else if res.StatusCode == 409 {
		return nil, ErrAlreadyInProgress
	}

	return nil, err
}

func (c *OtaApiClient) CancelOta(otaid string) (bool, error) {

	if otaid == "" {
		return false, fmt.Errorf("invalid ota-id: empty")
	}

	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return false, errors.New("wrong credentials")
		}
		return false, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	endpoint := c.host + "/ota/v1/ota/" + otaid + "/cancel"
	res, err := c.performRequest(endpoint, "PUT", userRequestToken.AccessToken)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	} else if res.StatusCode == 404 || res.StatusCode == 400 {
		return false, fmt.Errorf("ota-id %s not found", otaid)
	} else if res.StatusCode == 409 {
		return false, ErrAlreadyCancelled
	}

	return false, err
}
