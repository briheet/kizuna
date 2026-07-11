package config

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	Api APIConfig `mapstructure:",squash" validate:"required"`
	Db  DbConfig  `mapstructure:",squash" validate:"required"`
}

type APIConfig struct {
	Port              int `mapstructure:"port" validate:"required"`
	ReadHeaderTimeout int `mapstructure:"read_header_timeout" validate:"required"`
	ReadTimeout       int `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout      int `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout       int `mapstructure:"idle_timeout" validate:"required"`
}

type DbConfig struct {
	DatabaseURL string `mapstructure:"databaseurl" validate:"required"`
}

func LoadConfig(ctx context.Context, path string) (*Config, error) {

	var err error
	var config Config

	// Viper config
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("env")

	// If we have already injected in the environment
	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
