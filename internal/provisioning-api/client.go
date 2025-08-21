// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2025 ARDUINO SA (http://www.arduino.cc/)
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

package provisioningapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"golang.org/x/oauth2"
)

type ProvisioningApiClient struct {
	client       *http.Client
	host         string
	src          oauth2.TokenSource
	organization string
}

func NewClient(credentials *config.Credentials) *ProvisioningApiClient {
	host := iot.GetArduinoAPIBaseURL()
	tokenSource := iot.NewUserTokenSource(credentials.Client, credentials.Secret, host, credentials.Organization)
	return &ProvisioningApiClient{
		client:       &http.Client{},
		src:          tokenSource,
		host:         host,
		organization: credentials.Organization,
	}
}

func (c *ProvisioningApiClient) performRequest(endpoint, method, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, body)
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

func (c *ProvisioningApiClient) getToken() (*oauth2.Token, error) {
	token, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}
	return token, nil
}

func (c *ProvisioningApiClient) ClaimDevice(data ClaimData) (*ClaimResponse, *BadResponse, error) {
	endpoint := c.host + "provisioning/v1/onboarding/claim"
	token, err := c.getToken()
	if err != nil {
		return nil, nil, err
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal claim data: %w", err)
	}

	res, err := c.performRequest(endpoint, http.MethodPost, token.AccessToken, bytes.NewReader(dataJson))
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode == http.StatusOK {
		var response ClaimResponse

		err = json.Unmarshal(respBytes, &response)
		if err != nil {
			return nil, nil, err
		}
		return &response, nil, nil
	}
	var badResponse BadResponse

	err = json.Unmarshal(respBytes, &badResponse)
	if err != nil {
		return nil, nil, err
	}

	return nil, &badResponse, nil
}

func (c *ProvisioningApiClient) RegisterDevice(data RegisterBoardData) (*BadResponse, error) {
	endpoint := c.host + "provisioning/v1/boards/register"
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register data: %w", err)
	}

	res, err := c.performRequest(endpoint, http.MethodPost, token.AccessToken, bytes.NewReader(dataJson))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		return nil, nil
	}
	var badResponse BadResponse

	err = json.Unmarshal(respBytes, &badResponse)
	if err != nil {
		return nil, err
	}

	return &badResponse, nil
}

func (c *ProvisioningApiClient) Unclaim(provisioningId string) (*BadResponse, error) {
	endpoint := c.host + "provisioning/v1/onboarding/" + provisioningId
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	res, err := c.performRequest(endpoint, http.MethodDelete, token.AccessToken, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil, nil
	}
	var badResponse BadResponse
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respBytes, &badResponse)
	if err != nil {
		return nil, err
	}

	return &badResponse, nil
}

func (c *ProvisioningApiClient) GetProvisioningList() (*OnboardingsResponse, error) {
	endpoint := c.host + "provisioning/v1/onboarding?all=true"
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	res, err := c.performRequest(endpoint, http.MethodGet, token.AccessToken, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var response OnboardingsResponse

		respBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(respBytes, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	} else if res.StatusCode == 400 {
		return nil, errors.New(endpoint + " returned bad request")
	} else if res.StatusCode == 401 {
		return nil, errors.New(endpoint + " returned unauthorized request")
	} else if res.StatusCode == 403 {
		return nil, errors.New(endpoint + " returned forbidden request")
	} else if res.StatusCode == 500 {
		return nil, errors.New(endpoint + " returned internal server error")
	}

	return nil, err
}

func (c *ProvisioningApiClient) GetProvisioningDetail(provID string) (*Onboarding, error) {
	onboardingList, err := c.GetProvisioningList()
	if err != nil {
		return nil, fmt.Errorf("failed to get provisioning list: %w", err)
	}

	for _, onboarding := range onboardingList.Onboardings {
		if onboarding.ID == provID {
			return &onboarding, nil
		}
	}

	return nil, fmt.Errorf("onboarding with ID %s not found", provID)
}

func (c *ProvisioningApiClient) GetBoardsDetail() (*BoardTypeList, error) {
	endpoint := c.host + "iot/v1/supported/devices"
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	res, err := c.performRequest(endpoint, http.MethodGet, token.AccessToken, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var response BoardTypeList

		respBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(respBytes, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	} else if res.StatusCode == 400 {
		return nil, errors.New(endpoint + " returned bad request")
	} else if res.StatusCode == 401 {
		return nil, errors.New(endpoint + " returned unauthorized request")
	} else if res.StatusCode == 403 {
		return nil, errors.New(endpoint + " returned forbidden request")
	} else if res.StatusCode == 500 {
		return nil, errors.New(endpoint + " returned internal server error")
	}

	return nil, err
}
