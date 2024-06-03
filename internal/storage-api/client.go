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
	"strings"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"golang.org/x/oauth2"
)

const TemplateFileExtension = ".tino"

type StorageApiClient struct {
	client       *http.Client
	host         string
	src          oauth2.TokenSource
	organization string
}

func NewClient(credentials *config.Credentials) *StorageApiClient {
	host := iot.GetArduinoAPIBaseURL()
	tokenSource := iot.NewUserTokenSource(credentials.Client, credentials.Secret, host)
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

func (c *StorageApiClient) ExportCustomTemplate(templateId string) (*string, error) {

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

	if res.StatusCode == 200 {
		outfile, fileExportPath, err := createNewLocalFile(templateId, res)
		if err != nil {
			return nil, err
		}
		defer outfile.Close()
		io.Copy(outfile, res.Body)
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

func createNewLocalFile(templateId string, res *http.Response) (*os.File, string, error) {
	fileExportPath, err := composeNewLocalFileName(templateId, res)
	if err != nil {
		return nil, "", err
	}
	outfile, err := os.Create(fileExportPath)
	if err != nil {
		return nil, "", err
	}
	return outfile, fileExportPath, nil
}

func composeNewLocalFileName(templateId string, res *http.Response) (string, error) {
	fileExportPath := extractFileNameFromHeader(res)
	originalFileExportName := fileExportPath
	if fileExportPath == "" {
		fileExportPath = templateId + TemplateFileExtension
	}

	i := 1
	for ; i < 51; i++ {
		_, err := os.Stat(fileExportPath)
		if err != nil {
			if os.IsNotExist(err) {
				break
			} else {
				newbase := strings.TrimSuffix(originalFileExportName, TemplateFileExtension)
				newbase = newbase + "_" + string(i) + TemplateFileExtension
				fileExportPath = newbase
			}
		}
	}
	if i >= 50 {
		return "", errors.New("cannot create a new file name. Max number of copy reached.")
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
