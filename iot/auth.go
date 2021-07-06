package iot

import (
	"context"
	"net/url"

	"golang.org/x/oauth2"
	cc "golang.org/x/oauth2/clientcredentials"
)

func token(client, secret string) (*oauth2.Token, error) {
	// We need to pass the additional "audience" var to request an access token
	additionalValues := url.Values{}
	additionalValues.Add("audience", "https://api2.arduino.cc/iot")
	// Set up OAuth2 configuration
	config := cc.Config{
		ClientID:       client,
		ClientSecret:   secret,
		TokenURL:       "https://api2.arduino.cc/iot/v1/clients/token",
		EndpointParams: additionalValues,
	}
	// Get the access token in exchange of client_id and client_secret
	return config.Token(context.Background())
}
