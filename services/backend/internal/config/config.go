package config

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	Api      APIConfig      `mapstructure:",squash" validate:"required"`
	AI       AIConfig       `mapstructure:",squash" validate:"required"`
	Db       DbConfig       `mapstructure:",squash" validate:"required"`
	Embedder EmbedderConfig `mapstructure:",squash" validate:"required"`
}

type AIConfig struct {
	APIKey          string `mapstructure:"openai_api_key" validate:"required"`
	BaseURL         string `mapstructure:"ai_base_url" validate:"required,url"`
	Model           string `mapstructure:"ai_model" validate:"required"`
	MaxOutputTokens int    `mapstructure:"ai_max_output_tokens" validate:"required,min=1,max=4096"`
}

type APIConfig struct {
	Port              int    `mapstructure:"port" validate:"required"`
	CORSAllowedOrigin string `mapstructure:"cors_allowed_origin" validate:"required"`
	ReadHeaderTimeout int    `mapstructure:"read_header_timeout" validate:"required"`
	ReadTimeout       int    `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout      int    `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout       int    `mapstructure:"idle_timeout" validate:"required"`
}

type DbConfig struct {
	DatabaseURL string `mapstructure:"databaseurl" validate:"required"`
}

type EmbedderConfig struct {
	BaseURL string `mapstructure:"embedder_base_url" validate:"required,url"`
	Model   string `mapstructure:"embedder_model" validate:"required"`
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
