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

package storageapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const TemplateFileExtension = ".tino"

type StorageApiClient struct {
	client       *http.Client
	host         string
	src          oauth2.TokenSource
	organization string
}

func getArduinoAPIBaseURL() string {
	baseURL := "https://api-media.arduino.cc"
	if url := os.Getenv("IOT_API_MEDIA_URL"); url != "" {
		baseURL = url
	}
	return baseURL
}

func NewClient(credentials *config.Credentials) *StorageApiClient {
	host := getArduinoAPIBaseURL()
	iothost := iot.GetArduinoAPIBaseURL()
	tokenSource := iot.NewUserTokenSource(credentials.Client, credentials.Secret, iothost)
	return &StorageApiClient{
		client:       &http.Client{},
		src:          tokenSource,
		host:         host,
		organization: credentials.Organization,
	}
}

func (c *StorageApiClient) performMultipartRequest(endpoint, method, token, filename, multipartFieldName string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a form file
	formFile, err := writer.CreateFormFile(multipartFieldName, file.Name())
	if err != nil {
		return nil, err
	}

	// Copy the file data to the form file
	_, err = io.Copy(formFile, file)
	if err != nil {
		return nil, err
	}
	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, endpoint, &buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	if c.organization != "" {
		req.Header.Add("X-Organization", c.organization)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *StorageApiClient) performBinaryGetRequest(endpoint, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	if c.organization != "" {
		req.Header.Add("X-Organization", c.organization)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *StorageApiClient) performRequest(method, endpoint, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	if c.organization != "" {
		req.Header.Add("X-Organization", c.organization)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *StorageApiClient) ImportCustomTemplate(templateFile string) (*ImportCustomTemplateResponse, error) {

	if templateFile == "" {
		return nil, fmt.Errorf("invalid template: no file provided")
	}

	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	endpoint := c.host + "/storage/template/archive/v1/"
	res, err := c.performMultipartRequest(endpoint, "POST", userRequestToken.AccessToken, templateFile, "template")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyb, err := io.ReadAll(res.Body)

	if res.StatusCode == 200 {
		var importTemplResponse ImportCustomTemplateResponse
		if err == nil && bodyb != nil {
			err = json.Unmarshal(bodyb, &importTemplResponse)
			if err != nil {
				return nil, err
			}
		}

		return &importTemplResponse, nil
	} else if res.StatusCode == 400 {
		return nil, fmt.Errorf("bad request: %s", string(bodyb))
	} else if res.StatusCode == 409 {
		return nil, fmt.Errorf("template already exists: %s", string(bodyb))
	} else if res.StatusCode == 401 {
		return nil, errors.New("unauthorized request")
	} else if res.StatusCode == 403 {
		return nil, errors.New("forbidden request")
	} else if res.StatusCode == 500 {
		return nil, errors.New("internal server error")
	}

	return nil, err
}

func (c *StorageApiClient) ExportCustomTemplate(templateId, path string) (*string, error) {

	if templateId == "" {
		return nil, fmt.Errorf("invalid template id: no id provided")
	}

	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	endpoint := c.host + "/storage/template/archive/v1/" + templateId
	res, err := c.performBinaryGetRequest(endpoint, userRequestToken.AccessToken)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	logrus.Debugf("Export API call status: %d", res.StatusCode)

	if res.StatusCode == 200 || res.StatusCode == 201 {
		outfile, fileExportPath, err := createNewLocalFile(templateId, path, res)
		if err != nil {
			return nil, err
		}
		defer outfile.Close()
		_, err = io.Copy(outfile, res.Body)
		if err != nil {
			return nil, err
		}
		return &fileExportPath, nil
	} else if res.StatusCode == 400 {
		bodyb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad request: %s", string(bodyb))
	} else if res.StatusCode == 401 {
		return nil, errors.New("unauthorized request")
	} else if res.StatusCode == 403 {
		return nil, errors.New("forbidden request")
	} else if res.StatusCode == 500 {
		return nil, errors.New("internal server error")
	}

	return nil, err
}

func createNewLocalFile(templateId, path string, res *http.Response) (*os.File, string, error) {
	fileExportPath, err := composeNewLocalFileName(templateId, path, res)
	if err != nil {
		return nil, "", err
	}
	outfile, err := os.Create(fileExportPath)
	if err != nil {
		return nil, "", err
	}
	return outfile, fileExportPath, nil
}

func composeNewLocalFileName(templateId, path string, res *http.Response) (string, error) {
	fileExportPath := extractFileNameFromHeader(res)
	if path != "" {
		fileExportPath = filepath.Join(path, fileExportPath)
	}
	originalFileExportName := fileExportPath
	if fileExportPath == "" {
		fileExportPath = templateId + TemplateFileExtension
	}

	i := 1
	for ; i < 51; i++ {
		fileE, err := os.Stat(fileExportPath)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
		}
		if fileE != nil {
			newbase := strings.TrimSuffix(originalFileExportName, TemplateFileExtension)
			newbase = newbase + "_" + strconv.Itoa(i) + TemplateFileExtension
			fileExportPath = newbase
		}
	}
	if i >= 50 {
		return "", errors.New("cannot create a new file name. Max number of copy reached")
	}

	return fileExportPath, nil
}

func extractFileNameFromHeader(res *http.Response) string {
	content := res.Header.Get("Content-Disposition")
	if strings.HasPrefix(content, "attachment;") {
		content = strings.TrimPrefix(content, "attachment;")
		content = strings.TrimSpace(content)
		content = strings.TrimPrefix(content, "filename=")
		return strings.Trim(content, "\"")
	}
	return ""
}

type listPostRequest struct {
	Sort string `json:"sort"`
}

func (c *StorageApiClient) ListCustomTemplates() (*TemplatesListResponse, error) {
	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	request := listPostRequest{
		Sort: "asc",
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	endpoint := c.host + "/storage/template/v1/list"
	res, err := c.performRequest("POST", endpoint, userRequestToken.AccessToken, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	logrus.Debugf("List templates API call status: %d", res.StatusCode)

	if res.StatusCode == 200 || res.StatusCode == 201 {
		var templatesListResponse TemplatesListResponse
		respBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(respBytes, &templatesListResponse)
		if err != nil {
			return nil, err
		}
		return &templatesListResponse, nil
	} else if res.StatusCode == 400 {
		bodyb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad request: %s", string(bodyb))
	} else if res.StatusCode == 401 {
		return nil, errors.New("unauthorized request")
	} else if res.StatusCode == 403 {
		return nil, errors.New("forbidden request")
	} else if res.StatusCode == 500 {
		return nil, errors.New("internal server error")
	}

	return nil, err
}

func (c *StorageApiClient) GetCustomTemplate(templateID uuid.UUID) (*DescribeTemplateResponse, error) {
	userRequestToken, err := c.src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}

	endpoint := c.host + fmt.Sprintf("/storage/template/v1/%s", templateID.String())
	res, err := c.performRequest("GET", endpoint, userRequestToken.AccessToken, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	logrus.Debugf("Get template %s API call status: %d", templateID.String(), res.StatusCode)

	if res.StatusCode == 200 || res.StatusCode == 201 {
		var getTemplateResponse DescribeTemplateResponse
		respBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(respBytes, &getTemplateResponse)
		if err != nil {
			return nil, err
		}
		return &getTemplateResponse, nil
	} else if res.StatusCode == 400 {
		bodyb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad request: %s", string(bodyb))
	} else if res.StatusCode == 401 {
		return nil, errors.New("unauthorized request")
	} else if res.StatusCode == 403 {
		return nil, errors.New("forbidden request")
	} else if res.StatusCode == 500 {
		return nil, errors.New("internal server error")
	}

	return nil, err
}
