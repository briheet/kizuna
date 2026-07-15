package types

type NomicEmbeddingRequest struct {
	Model string   `json:"model" validate:"required"`
	Input []string `json:"input" validate:"required,min=1,dive,required"`
}

type NomicEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings" validate:"required,min=1,dive,min=1"`
}
