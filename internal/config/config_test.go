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

func TestRetrieve(t *testing.T) {
	var (
		validSecret     = "qaRZGEbnQNNvmaeTLqy8Bxs22wLZ6H7obIiNSveTLPdoQuylANnuy6WBOw16XoqH"
		validClient     = "CQ4iZ5sebOfhGRwUn3IV0r1YFMNrMTIx"
		validConfig     = &Config{validClient, validSecret}
		invalidConfig   = &Config{"", validSecret}
		clientEnv       = EnvPrefix + "_CLIENT"
		secretEnv       = EnvPrefix + "_SECRET"
		clientEnvBackup *string
		secretEnvBackup *string
	)

	// Preserve user environment variables when executing this test
	_ = func() {
		if c, ok := os.LookupEnv(clientEnv); ok {
			clientEnvBackup = &c
		}
		if s, ok := os.LookupEnv(secretEnv); ok {
			secretEnvBackup = &s
		}
	}
	_ = func() {
		if clientEnvBackup != nil {
			os.Setenv(clientEnv, *clientEnvBackup)
			clientEnvBackup = nil
		}
		if secretEnvBackup != nil {
			os.Setenv(secretEnv, *secretEnvBackup)
			secretEnvBackup = nil
		}
	}

	tests := []struct {
		name         string
		pre          func()
		post         func()
		wantedConfig *Config
		wantedErr    bool
	}{
		{
			name: "valid config written in env",
			pre: func() {
				// pushEnv()
				os.Setenv(clientEnv, validConfig.Client)
				os.Setenv(secretEnv, validConfig.Secret)
			},
			post: func() {
				os.Unsetenv(clientEnv)
				os.Unsetenv(secretEnv)
				// popEnv()
			},
			wantedConfig: validConfig,
			wantedErr:    false,
		},

		{
			name: "invalid config written in env",
			pre: func() {
				// pushEnv()
				os.Setenv(clientEnv, validConfig.Client)
				os.Setenv(secretEnv, "")
			},
			post: func() {
				os.Unsetenv(clientEnv)
				os.Unsetenv(secretEnv)
				// popEnv()
			},
			wantedConfig: nil,
			wantedErr:    true,
		},

		{
			name: "valid config written in parent of cwd",
			pre: func() {
				parent := "test-parent"
				cwd := "test-parent/test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid config in parent dir
				os.Chdir(parent)
				b, _ := json.Marshal(validConfig)
				os.WriteFile(Filename+".json", b, os.FileMode(0777))
				// Cwd has no config file
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
			name: "invalid config written in cwd, ignore config of parent dir",
			pre: func() {
				parent := "test-parent"
				cwd := "test-parent/test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid config in parent dir
				os.Chdir(parent)
				b, _ := json.Marshal(validConfig)
				os.WriteFile(Filename+".json", b, os.FileMode(0777))
				// Write invalid config in cwd
				os.Chdir("test-cwd")
				b, _ = json.Marshal(invalidConfig)
				os.WriteFile(Filename+".json", b, os.FileMode(0777))
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
			name: "invalid config written in env, ignore valid config of cwd",
			pre: func() {
				cwd := "test-cwd"
				os.MkdirAll(cwd, os.FileMode(0777))
				// Write valid config in cwd
				os.Chdir(cwd)
				b, _ := json.Marshal(validConfig)
				os.WriteFile(Filename+".json", b, os.FileMode(0777))
				// Write invalid config in env
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
			got, err := Retrieve()
			tt.post()

			if tt.wantedErr && err == nil {
				t.Errorf("Expected an error, but got nil")
			}
			if !tt.wantedErr && err != nil {
				t.Errorf("Expected nil error, but got: %v", err)
			}

			if !cmp.Equal(got, tt.wantedConfig) {
				t.Errorf("Wrong config received, diff:\n%s", cmp.Diff(tt.wantedConfig, got))
			}
		})
	}
}

func TestValidate(t *testing.T) {
	var (
		validSecret = "qaRZGEbnQNNvmaeTLqy8Bxs22wLZ6H7obIiNSveTLPdoQuylANnuy6WBOw16XoqH"
		validClient = "CQ4iZ5sebOfhGRwUn3IV0r1YFMNrMTIx"
	)
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "valid config",
			config: &Config{
				Client: validClient,
				Secret: validSecret,
			},
			valid: true,
		},
		{
			name:   "invalid client id",
			config: &Config{Client: "", Secret: validSecret},
			valid:  false,
		},
		{
			name:   "invalid client secret",
			config: &Config{Client: validClient, Secret: ""},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf(
					"Wrong validation, the config was correct but an error was received: \nconfig: %v\nerr: %v",
					tt.config,
					err,
				)
			}
			if !tt.valid && err == nil {
				t.Errorf(
					"Wrong validation, the config was invalid but no error was received: \nconfig: %v",
					tt.config,
				)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	var (
		validSecret = "qaRZGEbnQNNvmaeTLqy8Bxs22wLZ6H7obIiNSveTLPdoQuylANnuy6WBOw16XoqH"
		validClient = "CQ4iZ5sebOfhGRwUn3IV0r1YFMNrMTIx"
	)
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "empty config",
			config: &Config{Client: "", Secret: ""},
			want:   true,
		},
		{
			name:   "config without id",
			config: &Config{Client: "", Secret: validSecret},
			want:   false,
		},
		{
			name:   "config without secret",
			config: &Config{Client: validClient, Secret: ""},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEmpty()
			if got != tt.want {
				t.Errorf("Expected %v but got %v, with config: %v", tt.want, got, tt.config)
			}
		})
	}
}
