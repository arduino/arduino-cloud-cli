package config

import (
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/spf13/viper"
)

// Config accepts a configured viper instance and
// saves it into a specific config file.
func Config(conf *viper.Viper) error {

	// TODO: check if conf has the correct parameters
	// e.g. 'client' and 'secret'

	if err := conf.WriteConfigAs("config.yaml"); err != nil {
		feedback.Errorf("Cannot create config file: %v", err)
		return err
	}
	return nil
}
