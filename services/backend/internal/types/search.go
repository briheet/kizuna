package types

type SearchRequest struct {
	Query string `json:"query" validate:"required"`
	Limit int    `json:"limit" validate:"required,min=1,max=50"`
}

type SearchResult struct {
	ChunkID     string  `json:"chunk_id"`
	GraphNodeID string  `json:"graph_node_id"`
	Content     string  `json:"content"`
	Title       string  `json:"title"`
	SourceLink  string  `json:"source_link"`
	SourceType  string  `json:"source_type"`
	NodeType    string  `json:"node_type"`
	Distance    float64 `json:"distance"`
}

type SearchResponse struct {
	Summary      string         `json:"summary"`
	Results      []SearchResult `json:"results"`
	RelatedNodes []RelatedNode  `json:"related_nodes"`
	Edges        []SearchEdge   `json:"edges"`
}

type RelatedNode struct {
	ID         string `json:"id"`
	NodeType   string `json:"node_type"`
	Title      string `json:"title"`
	SourceLink string `json:"source_link"`
}

type SearchEdge struct {
	ID            string  `json:"id"`
	FromNodeID    string  `json:"from_node_id"`
	ToNodeID      string  `json:"to_node_id"`
	EdgeType      string  `json:"edge_type"`
	EdgeScope     string  `json:"edge_scope"`
	Confidence    float64 `json:"confidence"`
	EvidenceNode  string  `json:"evidence_node_id"`
	EvidenceChunk string  `json:"evidence_chunk_id"`
}

type NomicEmbeddingRequest struct {
	Model string   `json:"model" validate:"required"`
	Input []string `json:"input" validate:"required,min=1,dive,required"`
}

type NomicEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings" validate:"required,min=1,dive,min=1"`
}
