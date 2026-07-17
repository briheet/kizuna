package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigUsesOpenAIKeyFromEnvironment(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "environment-api-key")

	path := filepath.Join(t.TempDir(), "backend.env")
	require.NoError(t, os.WriteFile(path, []byte(`
PORT=4000
CORS_ALLOWED_ORIGIN=*
READ_HEADER_TIMEOUT=5
READ_TIMEOUT=30
WRITE_TIMEOUT=60
IDLE_TIMEOUT=120
DATABASEURL=postgresql://root@127.0.0.1:26257/kizuna?sslmode=disable
EMBEDDER_BASE_URL=http://127.0.0.1:11434
EMBEDDER_MODEL=nomic-embed-text:v1.5
OPENAI_API_KEY=
AI_BASE_URL=https://api.openai.com
AI_MODEL=gpt-5.4-mini
AI_MAX_OUTPUT_TOKENS=700
`), 0o600))

	cfg, err := LoadConfig(t.Context(), path)
	require.NoError(t, err)
	require.Equal(t, "environment-api-key", cfg.AI.APIKey)
}
