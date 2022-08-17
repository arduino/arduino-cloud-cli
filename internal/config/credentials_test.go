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

package config

import (
	"os"
	"testing"

	"encoding/json"

	"github.com/google/go-cmp/cmp"
)

func TestRetrieveCredentials(t *testing.T) {
	var (
		validSecret             = "qaRZGEbnQNNvmaeTLqy8Bxs22wLZ6H7obIiNSveTLPdoQuylANnuy6WBOw16XoqH"
		validClient             = "CQ4iZ5sebOfhGRwUn3IV0r1YFMNrMTIx"
		validOrganization       = "dc6a6159-3cd5-41a2-b391-553b1351cd98"
		validConfig             = &Credentials{Client: validClient, Secret: validSecret}
		validWithOptionalConfig = &Credentials{Client: validClient, Secret: validSecret, Organization: validOrganization}
		invalidConfig           = &Credentials{Client: "", Secret: validSecret}
		clientEnv               = EnvPrefix + "_CLIENT"
		secretEnv               = EnvPrefix + "_SECRET"
		organizationEnv         = EnvPrefix + "_ORGANIZATION"
	)

	tests := []struct {
		name         string
		pre          func()
		post         func()
		wantedConfig *Credentials
		wantedErr    bool
	}{
		{
			name: "valid credentials with only mandatory params written in env",
			pre: func() {
				os.Setenv(clientEnv, validConfig.Client)
				os.Setenv(secretEnv, validConfig.Secret)
			},
			post: func() {
				os.Unsetenv(clientEnv)
				os.Unsetenv(secretEnv)
			},
			wantedConfig: validConfig,
			wantedErr:    false,
		},

		{
			name: "valid credentials with optional params written in env",
			pre: func() {
				os.Setenv(clientEnv, validWithOptionalConfig.Client)
				os.Setenv(secretEnv, validWithOptionalConfig.Secret)
				os.Setenv(organizationEnv, validWithOptionalConfig.Organization)
			},
			post: func() {
				os.Unsetenv(clientEnv)
				os.Unsetenv(secretEnv)
				os.Unsetenv(organizationEnv)
			},
			wantedConfig: validWithOptionalConfig,
			wantedErr:    false,
		},

		{
			name: "invalid credentials written in env",
			pre: func() {
				os.Setenv(clientEnv, validConfig.Client)
				os.Setenv(secretEnv, "")
			},
			post: func() {
				os.Unsetenv(clientEnv)
				os.Unsetenv(secretEnv)
			},
			wantedConfig: nil,
			wantedErr:    true,
		},

		{
			name: "valid credentials written in parent of cwd",
			pre: func() {
				parent := "test-parent"
				cwd := "test-parent/test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid credentials in parent dir
				os.Chdir(parent)
				b, _ := json.Marshal(validConfig)
				os.WriteFile(CredentialsFilename+".json", b, os.FileMode(0777))
				// Cwd has no credentials file
				os.Chdir("test-cwd")
			},
			post: func() {
				os.Chdir("../..")
				os.RemoveAll("test-parent")
			},
			wantedConfig: validConfig,
			wantedErr:    false,
		},

		{
			name: "valid credentials with optional params written in parent of cwd",
			pre: func() {
				parent := "test-parent"
				cwd := "test-parent/test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid credentials in parent dir
				os.Chdir(parent)
				b, _ := json.Marshal(validWithOptionalConfig)
				os.WriteFile(CredentialsFilename+".json", b, os.FileMode(0777))
				// Cwd has no credentials file
				os.Chdir("test-cwd")
			},
			post: func() {
				os.Chdir("../..")
				os.RemoveAll("test-parent")
			},
			wantedConfig: validWithOptionalConfig,
			wantedErr:    false,
		},

		{
			name: "invalid credentials written in cwd, ignore credentials of parent dir",
			pre: func() {
				parent := "test-parent"
				cwd := "test-parent/test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid credentials in parent dir
				os.Chdir(parent)
				b, _ := json.Marshal(validConfig)
				os.WriteFile(CredentialsFilename+".json", b, os.FileMode(0777))
				// Write invalid credentials in cwd
				os.Chdir("test-cwd")
				b, _ = json.Marshal(invalidConfig)
				os.WriteFile(CredentialsFilename+".json", b, os.FileMode(0777))
			},
			post: func() {
				os.Chdir("../..")
				os.RemoveAll("test-parent")
				os.Unsetenv(clientEnv)
				os.Unsetenv(secretEnv)
			},
			wantedConfig: nil,
			wantedErr:    true,
		},

		{
			name: "invalid credentials written in env, ignore valid credentials of cwd",
			pre: func() {
				cwd := "test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid credentials in cwd
				os.Chdir(cwd)
				b, _ := json.Marshal(validConfig)
				os.WriteFile(CredentialsFilename+".json", b, os.FileMode(0777))
				// Write invalid credentials in env
				os.Setenv(clientEnv, validConfig.Client)
				os.Setenv(secretEnv, "")
			},
			post: func() {
				os.Chdir("..")
				os.RemoveAll("test-cwd")
			},
			wantedConfig: nil,
			wantedErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pre()
			got, err := RetrieveCredentials()
			tt.post()

			if tt.wantedErr && err == nil {
				t.Errorf("Expected an error, but got nil")
			}
			if !tt.wantedErr && err != nil {
				t.Errorf("Expected nil error, but got: %v", err)
			}

			if !cmp.Equal(got, tt.wantedConfig) {
				t.Errorf("Wrong credentials received, diff:\n%s", cmp.Diff(tt.wantedConfig, got))
			}
		})
	}
}

func TestValidate(t *testing.T) {
	var (
		validSecret       = "qaRZGEbnQNNvmaeTLqy8Bxs22wLZ6H7obIiNSveTLPdoQuylANnuy6WBOw16XoqH"
		validClient       = "CQ4iZ5sebOfhGRwUn3IV0r1YFMNrMTIx"
		validOrganization = "dc6a6159-3cd5-41a2-b391-553b1351cd98"
	)
	tests := []struct {
		name   string
		config *Credentials
		valid  bool
	}{
		{
			name:   "valid credentials",
			config: &Credentials{Client: validClient, Secret: validSecret, Organization: validOrganization},
			valid:  true,
		},
		{
			name:   "valid credentials, organization is optional",
			config: &Credentials{Client: validClient, Secret: validSecret, Organization: ""},
			valid:  true,
		},
		{
			name:   "invalid client id",
			config: &Credentials{Client: "", Secret: validSecret},
			valid:  false,
		},
		{
			name:   "invalid client secret",
			config: &Credentials{Client: validClient, Secret: ""},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf(
					"Wrong validation, the credentials were correct but an error was received: \ncredentials: %v\nerr: %v",
					tt.config,
					err,
				)
			}
			if !tt.valid && err == nil {
				t.Errorf(
					"Wrong validation, the credentials were invalid but no error was received: \ncredentials: %v",
					tt.config,
				)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	var (
		validSecret       = "qaRZGEbnQNNvmaeTLqy8Bxs22wLZ6H7obIiNSveTLPdoQuylANnuy6WBOw16XoqH"
		validClient       = "CQ4iZ5sebOfhGRwUn3IV0r1YFMNrMTIx"
		validOrganization = "dc6a6159-3cd5-41a2-b391-553b1351cd98"
	)
	tests := []struct {
		name   string
		config *Credentials
		want   bool
	}{
		{
			name:   "empty credentials",
			config: &Credentials{Client: "", Secret: "", Organization: ""},
			want:   true,
		},
		{
			name:   "empty mandatory credentials - optionals given",
			config: &Credentials{Client: "", Secret: "", Organization: validOrganization},
			want:   true,
		},
		{
			name:   "credentials without id",
			config: &Credentials{Client: "", Secret: validSecret},
			want:   false,
		},
		{
			name:   "credentials without secret",
			config: &Credentials{Client: validClient, Secret: ""},
			want:   false,
		},
		{
			name:   "credentials with all mandatory params set",
			config: &Credentials{Client: validClient, Secret: validSecret},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEmpty()
			if got != tt.want {
				t.Errorf("Expected %v but got %v, with credentials: %v", tt.want, got, tt.config)
			}
		})
	}
}
