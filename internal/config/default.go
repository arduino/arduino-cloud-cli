package config

import "github.com/spf13/viper"

var (
	Filename = "arduino-cloud"
)

// SetDefaults sets the default values for configuration keys
func SetDefaults(settings *viper.Viper) {
	// Client ID
	settings.SetDefault("client", "xxxxxxxxxxxxxx")
	// Secret
	settings.SetDefault("secret", "xxxxxxxxxxxxxx")
}
