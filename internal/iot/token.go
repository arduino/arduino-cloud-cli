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
	"errors"
	"fmt"
	"net/url"
	"strings"

	iotclient "github.com/arduino/iot-client-go"
	"golang.org/x/oauth2"
	cc "golang.org/x/oauth2/clientcredentials"
)

func token(client, secret, baseURL string) oauth2.TokenSource {
	// We need to pass the additional "audience" var to request an access token.
	additionalValues := url.Values{}
	additionalValues.Add("audience", "https://api2.arduino.cc/iot")
	// Set up OAuth2 configuration.
	config := cc.Config{
		ClientID:       client,
		ClientSecret:   secret,
		TokenURL:       baseURL + "/iot/v1/clients/token",
		EndpointParams: additionalValues,
	}

	// Retrieve a token source that allows to retrieve tokens
	// with an automatic refresh mechanism.
	return config.TokenSource(context.Background())
}

func ctxWithToken(ctx context.Context, src oauth2.TokenSource) (context.Context, error) {
	// Retrieve a valid token from the src.
	tok, err := src.Token()
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return nil, errors.New("wrong credentials")
		}
		return nil, fmt.Errorf("cannot retrieve a valid token: %w", err)
	}
	return context.WithValue(ctx, iotclient.ContextAccessToken, tok.AccessToken), nil
}
