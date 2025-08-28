package iotapiraw

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/go-paths-helper"
	"golang.org/x/oauth2"
)

type IoTApiRawClient struct {
	client       *http.Client
	host         string
	src          oauth2.TokenSource
	organization string
}

func NewClient(credentials *config.Credentials) *IoTApiRawClient {
	host := iot.GetArduinoAPIBaseURL()
	tokenSource := iot.NewUserTokenSource(credentials.Client, credentials.Secret, host, credentials.Organization)
	return &IoTApiRawClient{
		client:       &http.Client{},
		src:          tokenSource,
		host:         host,
		organization: credentials.Organization,
	}
}

func (c *IoTApiRawClient) performRequest(endpoint, method, token string, body io.Reader) (*http.Response, error) {
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

func (c *IoTApiRawClient) getToken() (*oauth2.Token, error) {
	token, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}
	return token, nil
}

func (c *IoTApiRawClient) GetBoardsDetail() (*BoardTypeList, error) {
	endpoint := c.host + "/iot/v1/supported/devices"
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

func (c *IoTApiRawClient) GetBoardDetailByFQBN(fqbn string) (*BoardType, error) {
	boardsList, err := c.GetBoardsDetail()
	if err != nil {
		return nil, err
	}

	for _, b := range *boardsList {
		if b.FQBN != nil && *b.FQBN == fqbn {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("board with fqbn %s not found", fqbn)
}

func (c *IoTApiRawClient) DownloadProvisioningV2Sketch(fqbn string, path *paths.Path, filename *string) (string, error) {
	endpoint := c.host + "/iot/v2/binaries/provisioningv2?fqbn=" + fqbn
	token, err := c.getToken()
	if err != nil {
		return "", err
	}

	res, err := c.performRequest(endpoint, http.MethodGet, token.AccessToken, nil)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var response Prov2SketchBinRes
		respBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(respBytes, &response)
		if err != nil {
			return "", err
		}

		if filename != nil {
			path = path.Join(*filename)
		} else {
			path = path.Join(response.FileName)
		}

		path.Parent().MkdirAll()

		bytes, err := base64.StdEncoding.DecodeString(response.Binary)
		if err != nil {
			return "", err
		}

		if err = path.WriteFile(bytes); err != nil {
			return "", fmt.Errorf("writing provisioning v2 binary: %w", err)
		}
		p, err := path.Abs()
		if err != nil {
			return "", fmt.Errorf("cannot retrieve absolute path of downloaded provisioning v2 binary: %w", err)
		}
		return p.String(), nil
	} else if res.StatusCode == 400 {
		return "", errors.New(endpoint + " returned bad request")
	} else if res.StatusCode == 401 {
		return "", errors.New(endpoint + " returned unauthorized request")
	} else if res.StatusCode == 403 {
		return "", errors.New(endpoint + " returned forbidden request")
	} else if res.StatusCode == 500 {
		return "", errors.New(endpoint + " returned internal server error")
	}

	return "", errors.New("failed to download the provisioning v2 binary: unknown error")

}
