package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigAllowsUnusedProvidersToBeEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "workers.env")
	err := os.WriteFile(path, []byte(`
DATABASEURL=postgresql://root@127.0.0.1:26257/kizuna?sslmode=disable
EMBEDDER_BASE_URL=http://127.0.0.1:11434
EMBEDDER_MODEL=nomic-embed-text:v1.5
GITHUB_TOKEN=github-token
CONFLUENCE_HOST=
CONFLUENCE_MAIL=
CONFLUENCE_TOKEN=
DISCORD_TOKEN=
DISCORD_TOKEN_TYPE=
GITHUB_TOKEN_TYPE=
SLACK_TOKEN=
JIRA_HOST=
JIRA_MAIL=
JIRA_TOKEN=
`), 0o600)
	require.NoError(t, err)

	cfg, err := LoadConfig(t.Context(), path)
	require.NoError(t, err)
	require.Empty(t, cfg.Confluence.Token)
	require.Empty(t, cfg.Discord.Token)
	require.Empty(t, cfg.Slack.Token)
	require.Empty(t, cfg.Jira.Token)
}

func TestLoadConfigMergesGitHubCredentialFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "workers.env")
	require.NoError(t, os.WriteFile(path, []byte(`
DATABASEURL=postgresql://root@127.0.0.1:26257/kizuna?sslmode=disable
EMBEDDER_BASE_URL=http://127.0.0.1:11434
EMBEDDER_MODEL=nomic-embed-text:v1.5
GITHUB_TOKEN=
`), 0o600))
	credentialPath := filepath.Join(directory, "workers-secrets.env")
	require.NoError(t, os.WriteFile(credentialPath, []byte("GITHUB_TOKEN=credential-github-token\n"), 0o600))

	cfg, err := LoadConfig(t.Context(), path, credentialPath)
	require.NoError(t, err)
	require.Equal(t, "credential-github-token", cfg.Github.Token)
}
