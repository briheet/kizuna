package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/briheet/kizuna/workers/internal/types"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type NomicRepository struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewNomicRepository(baseURL string) *NomicRepository {
	return &NomicRepository{
		baseURL: baseURL,
		model:   "nomic-embed-text",
		client:  http.DefaultClient,
	}
}

func (r *NomicRepository) Dimensions() int {
	return 768
}

func (r *NomicRepository) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	input := types.NomicEmbeddingRequest{
		Model: r.model,
		Input: texts,
	}
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("nomic embed failed: %s", resp.Status)
	}

	var out types.NomicEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if err := validate.Struct(out); err != nil {
		return nil, err
	}

	return out.Embeddings, nil
}
