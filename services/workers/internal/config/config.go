package config

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	Confluence ConfluenceConfig `mapstructure:",squash" validate:"required"`
	Db         DbConfig         `mapstructure:",squash" validate:"required"`
	Discord    DiscordConfig    `mapstructure:",squash" validate:"required"`
	Embedder   EmbedderConfig   `mapstructure:",squash" validate:"required"`
	Github     GithubConfig     `mapstructure:",squash" validate:"required"`
	Slack      SlackConfig      `mapstructure:",squash" validate:"required"`
	Jira       JiraConfig       `mapstructure:",squash" validate:"required"`
}

type ConfluenceConfig struct {
	Host  string `mapstructure:"confluence_host" validate:"required"`
	Mail  string `mapstructure:"confluence_mail" validate:"required"`
	Token string `mapstructure:"confluence_token" validate:"required"`
}

type DbConfig struct {
	DatabaseURL string `mapstructure:"databaseurl" validate:"required"`
}

type DiscordConfig struct {
	Token     string `mapstructure:"discord_token" validate:"required"`
	TokenType string `mapstructure:"discord_token_type" validate:"required"`
}

type GithubConfig struct {
	Token     string `mapstructure:"github_token" validate:"required"`
	TokenType string `mapstructure:"github_token_type" validate:"required"`
}

type EmbedderConfig struct {
	BaseURL string `mapstructure:"embedder_base_url" validate:"required,url"`
}

type SlackConfig struct {
	Token string `mapstructure:"slack_token" validate:"required"`
}

type JiraConfig struct {
	Host  string `mapstructure:"jira_host" validate:"required"`
	Mail  string `mapstructure:"jira_mail" validate:"required"`
	Token string `mapstructure:"jira_token" validate:"required"`
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
