package cockroachdb

import (
	"context"
	"strconv"
	"strings"

	"github.com/briheet/kizuna/backend/internal/db"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CockroachDbSearchRepository struct {
	db *db.Client
}

func NewCockroachDbSearchRepository(db *db.Client) *CockroachDbSearchRepository {
	return &CockroachDbSearchRepository{db: db}
}

func (r *CockroachDbSearchRepository) SearchChunks(ctx context.Context, topicID uuid.UUID, embedding []float32, limit int) ([]types.SearchResult, error) {
	var results []types.SearchResult

	err := r.db.ExecuteTx(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			select
				c.id::string,
				c.graph_node_id::string,
				c.content,
				coalesce(gn.title, ''),
				coalesce(gn.source_link, ''),
				c.embedding <=> $1::VECTOR as distance
			from chunks c
			join graph_nodes gn on gn.id = c.graph_node_id
			join data_sources ds on ds.id = gn.data_source_id
			where ds.topic_id = $2
			  and c.embedding is not null
			order by c.embedding <=> $1::VECTOR
			limit $3;
		`, vectorLiteral(embedding), topicID, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var result types.SearchResult
			if err := rows.Scan(&result.ChunkID, &result.GraphNodeID, &result.Content, &result.Title, &result.SourceLink, &result.Distance); err != nil {
				return err
			}
			results = append(results, result)
		}

		return rows.Err()
	})

	return results, err
}

func (r *CockroachDbSearchRepository) GetRelatedGraph(ctx context.Context, nodeIDs []uuid.UUID) ([]types.RelatedNode, []types.SearchEdge, error) {
	if len(nodeIDs) == 0 {
		return nil, nil, nil
	}

	var nodes []types.RelatedNode
	var edges []types.SearchEdge

	err := r.db.ExecuteTx(ctx, func(tx pgx.Tx) error {
		edgeRows, err := tx.Query(ctx, `
			select
				id::string,
				from_node_id::string,
				to_node_id::string,
				edge_type,
				edge_scope,
				confidence,
				coalesce(evidence_node_id::string, ''),
				coalesce(evidence_chunk_id::string, '')
			from graph_edges
			where from_node_id = any($1)
			   or to_node_id = any($1);
		`, nodeIDs)
		if err != nil {
			return err
		}
		defer edgeRows.Close()

		for edgeRows.Next() {
			var edge types.SearchEdge
			if err := edgeRows.Scan(&edge.ID, &edge.FromNodeID, &edge.ToNodeID, &edge.EdgeType, &edge.EdgeScope, &edge.Confidence, &edge.EvidenceNode, &edge.EvidenceChunk); err != nil {
				return err
			}
			edges = append(edges, edge)
		}
		if err := edgeRows.Err(); err != nil {
			return err
		}

		nodeRows, err := tx.Query(ctx, `
			with related as (
				select unnest($1::uuid[]) as id
				union
				select from_node_id from graph_edges where to_node_id = any($1)
				union
				select to_node_id from graph_edges where from_node_id = any($1)
			)
			select
				gn.id::string,
				gn.node_type,
				coalesce(gn.title, ''),
				coalesce(gn.source_link, '')
			from graph_nodes gn
			join related r on r.id = gn.id;
		`, nodeIDs)
		if err != nil {
			return err
		}
		defer nodeRows.Close()

		for nodeRows.Next() {
			var node types.RelatedNode
			if err := nodeRows.Scan(&node.ID, &node.NodeType, &node.Title, &node.SourceLink); err != nil {
				return err
			}
			nodes = append(nodes, node)
		}

		return nodeRows.Err()
	})

	return nodes, edges, err
}

func vectorLiteral(values []float32) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
