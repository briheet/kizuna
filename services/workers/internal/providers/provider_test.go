package providers

import (
	"testing"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/stretchr/testify/require"
)

func TestEnabledProvidersSkipsUnconfiguredProviders(t *testing.T) {
	cfg := &config.Config{
		Github: config.GithubConfig{Token: "github-token"},
		Slack:  config.SlackConfig{Token: "slack-token"},
	}

	require.Equal(t, []string{"github", "slack"}, EnabledProviders(cfg))
}
