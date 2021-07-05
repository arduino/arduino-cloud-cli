package config

import (
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/spf13/viper"
)

type Config struct {
	Client string `yaml:"client"`
	Secret string `yaml:"secret"`
}

func Retrieve() (*Config, error) {
	conf := &Config{}
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		feedback.Errorf("Fatal error config file:  %v", err)
		return nil, err
	}

	v.Unmarshal(conf)
	return conf, nil
}
