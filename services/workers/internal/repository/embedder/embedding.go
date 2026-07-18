package embedder

import (
	"context"
	"fmt"
	"strings"

	embedderclient "github.com/briheet/kizuna/workers/internal/embedder"
)

const nomicDimensions = 768

type NomicRepository struct {
	client *embedderclient.Client
	model  string
}

type nomicEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type nomicEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func NewNomicRepository(client *embedderclient.Client, model string) *NomicRepository {
	return &NomicRepository{client: client, model: model}
}

func (r *NomicRepository) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if err := validateInputs(texts); err != nil {
		return nil, err
	}

	inputs := make([]string, len(texts))
	for index, text := range texts {
		inputs[index] = "search_document: " + text
	}

	return r.embed(ctx, inputs)
}

func (r *NomicRepository) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if err := validateInputs([]string{text}); err != nil {
		return nil, err
	}

	vectors, err := r.embed(ctx, []string{"search_query: " + text})
	if err != nil {
		return nil, err
	}

	return vectors[0], nil
}

func (r *NomicRepository) Dimensions() int {
	return nomicDimensions
}

func (r *NomicRepository) embed(ctx context.Context, texts []string) ([][]float32, error) {
	var response nomicEmbeddingResponse
	err := r.client.PostJSON(ctx, "/api/embed", nomicEmbeddingRequest{
		Model: r.model,
		Input: texts,
	}, &response)
	if err != nil {
		return nil, fmt.Errorf("nomic embed: %w", err)
	}

	if len(response.Embeddings) != len(texts) {
		return nil, fmt.Errorf("nomic embed: expected %d embeddings, got %d", len(texts), len(response.Embeddings))
	}
	for index, vector := range response.Embeddings {
		if len(vector) != nomicDimensions {
			return nil, fmt.Errorf("nomic embed: embedding %d has %d dimensions, expected %d", index, len(vector), nomicDimensions)
		}
	}

	return response.Embeddings, nil
}

func validateInputs(texts []string) error {
	if len(texts) == 0 {
		return fmt.Errorf("nomic embed: input is empty")
	}
	for index, text := range texts {
		if strings.TrimSpace(text) == "" {
			return fmt.Errorf("nomic embed: input %d is empty", index)
		}
	}

	return nil
}
