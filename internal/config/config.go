package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config contains all the configuration parameters
// known by arduino-cloud-cli
type Config struct {
	Client string `map-structure:"client"` // Client ID of the user
	Secret string `map-structure:"secret"` // Secret ID of the user, unique for each Client ID
}

// Retrieve returns the actual parameters contained in the
// configuration file, if any. Returns error if no config file is found.
func Retrieve() (*Config, error) {
	conf := &Config{}
	v := viper.New()
	v.SetConfigName(Filename)
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("%s: %w", "retrieving config file", err)
		return nil, err
	}

	v.Unmarshal(conf)
	return conf, nil
}
