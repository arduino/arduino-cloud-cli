package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arduino/go-paths-helper"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/spf13/viper"
)

// InitParams contains the parameters needed to initialize a configuration file.
type InitParams struct {
	DestDir   string // Destination directory in which the configuration file will be saved
	Overwrite bool   // Overwrite specifies if existing config file should be overwritten
	Format    string // Config file format, can be 'json' or 'yaml'
}

func validateFormatString(arg string) error {
	if arg != "json" && arg != "yaml" {
		return errors.New("passed format is not valid, select between 'json' and 'yaml'")
	}
	return nil
}

// Init initializes a configuration file with default values.
// If the file doesn't exist, it is created.
// If it exists, it is written to only if overwrite param is true.
func Init(params *InitParams) error {
	configPath, err := paths.New(params.DestDir).Abs()
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot retrieve absolute path of passed dest-dir", err)
	}
	if !configPath.IsDir() {
		return fmt.Errorf("%s: %w", "passed dest-dir is not a valid directory", err)
	}

	params.Format = strings.ToLower(params.Format)
	if err := validateFormatString(params.Format); err != nil {
		return err
	}

	configFile := configPath.Join(config.Filename + "." + params.Format)

	if !params.Overwrite && configFile.Exist() {
		return errors.New("config file already exists, use --overwrite to discard the existing one")
	}

	newSettings := viper.New()
	config.SetDefaults(newSettings)
	if err := newSettings.WriteConfigAs(configFile.String()); err != nil {
		return fmt.Errorf("cannot create config file: %v", err)
	}

	return nil
}
