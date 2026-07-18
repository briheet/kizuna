package cockroachdb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GraphRepository struct {
	db *db.Client
}

func NewGraphRepository(db *db.Client) *GraphRepository {
	return &GraphRepository{db: db}
}

func (r *GraphRepository) UpsertDataSource(ctx context.Context, input repository.DataSourceInput) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.ExecuteTx(ctx, func(tx pgx.Tx) error {
		id = uuid.New()
		err := tx.QueryRow(ctx, `
			insert into data_sources (
				id, topic_id, source_type, name, external_id, source_link,
				config, created_at, updated_at
			) values ($1, $2, $3, $4, $5, $6, $7, now(), now())
			on conflict (topic_id, source_type, external_id)
			do update set
				name = excluded.name,
				source_link = excluded.source_link,
				config = excluded.config,
				updated_at = now()
			returning id;
		`, id, input.TopicID, input.SourceType, input.Name, input.ExternalID, input.SourceLink, input.Config).Scan(&id)
		if err != nil {
			return fmt.Errorf("upsert data source: %w", err)
		}
		return nil
	})

	return id, err
}

func (r *GraphRepository) SaveGraph(ctx context.Context, dataSourceID uuid.UUID, graph repository.GraphInput) error {
	return r.db.ExecuteTx(ctx, func(tx pgx.Tx) error {
		nodeIDs := map[string]uuid.UUID{}

		for _, input := range graph.Nodes {
			nodeID, err := upsertGraphNode(ctx, tx, dataSourceID, input.Node)
			if err != nil {
				return err
			}
			nodeIDs[input.Node.ExternalID] = nodeID

			for _, chunk := range input.Chunks {
				if chunk.Content == "" {
					continue
				}
				if _, err := tx.Exec(ctx, `
					insert into chunks (id, graph_node_id, chunk_index, content, embedding, created_at)
					values ($1, $2, $3, $4, $5::VECTOR, now())
					on conflict (graph_node_id, chunk_index)
					do update set
						content = excluded.content,
						embedding = excluded.embedding;
				`, uuid.New(), nodeID, chunk.Index, chunk.Content, vectorLiteral(chunk.Embedding)); err != nil {
					return err
				}
			}
		}

		for _, edge := range graph.Edges {
			if _, err := tx.Exec(ctx, `
				insert into graph_edges (
					id, root_data_source_id, from_node_id, to_node_id, edge_type,
					edge_scope, confidence, properties, created_at
				) values ($1, $2, $3, $4, $5, $6, $7, $8, now())
				on conflict (root_data_source_id, from_node_id, edge_type, to_node_id)
				do update set
					edge_scope = excluded.edge_scope,
					confidence = excluded.confidence,
					properties = excluded.properties;
			`, uuid.New(), dataSourceID, nodeIDs[edge.FromExternalID], nodeIDs[edge.ToExternalID], edge.EdgeType, edge.EdgeScope, edge.Confidence, edge.Properties); err != nil {
				return err
			}
		}

		return nil
	})
}

func upsertGraphNode(ctx context.Context, tx pgx.Tx, dataSourceID uuid.UUID, node repository.GraphNodeInput) (uuid.UUID, error) {
	id := uuid.New()
	err := tx.QueryRow(ctx, `
		insert into graph_nodes (
			id, data_source_id, node_type, external_id, source_link,
			title, path, properties, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		on conflict (data_source_id, node_type, external_id)
		do update set
			source_link = excluded.source_link,
			title = excluded.title,
			path = excluded.path,
			properties = excluded.properties,
			updated_at = now()
		returning id;
	`, id, dataSourceID, node.NodeType, node.ExternalID, node.SourceLink, node.Title, node.Path, node.Properties).Scan(&id)
	return id, err
}

func vectorLiteral(values []float32) any {
	if len(values) == 0 {
		return nil
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
