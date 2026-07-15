CREATE TABLE "organisations" (
  "id" uuid PRIMARY KEY,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL
);

CREATE TABLE "teams" (
  "id" uuid PRIMARY KEY,
  "organisation_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL
);

CREATE TABLE "topics" (
  "id" uuid PRIMARY KEY,
  "team_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL
);

CREATE TABLE "data_sources" (
  "id" uuid PRIMARY KEY,
  "topic_id" uuid NOT NULL,
  "source_type" varchar NOT NULL,
  "name" varchar NOT NULL,
  "external_id" varchar,
  "source_link" varchar,
  "config" jsonb,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL
);

CREATE UNIQUE INDEX "data_sources_unique_source" ON "data_sources" ("topic_id", "source_type", "external_id");

CREATE TABLE "graph_nodes" (
  "id" uuid PRIMARY KEY,
  "data_source_id" uuid NOT NULL,
  "node_type" varchar NOT NULL,
  "external_id" varchar NOT NULL,
  "source_link" varchar,
  "title" varchar,
  "path" varchar,
  "properties" jsonb,
  "source_created_at" timestamptz,
  "source_updated_at" timestamptz,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL
);

CREATE UNIQUE INDEX "graph_nodes_unique_external" ON "graph_nodes" ("data_source_id", "node_type", "external_id");

CREATE TABLE "graph_edges" (
  "id" uuid PRIMARY KEY,
  "root_data_source_id" uuid NOT NULL,
  "from_node_id" uuid NOT NULL,
  "to_node_id" uuid NOT NULL,
  "edge_type" varchar NOT NULL,
  "edge_scope" varchar NOT NULL,
  "confidence" float NOT NULL,
  "properties" jsonb,
  "evidence_node_id" uuid,
  "evidence_chunk_id" uuid,
  "created_at" timestamptz NOT NULL
);

CREATE UNIQUE INDEX "graph_edges_unique_relation" ON "graph_edges" ("root_data_source_id", "from_node_id", "edge_type", "to_node_id");

CREATE TABLE "chunks" (
  "id" uuid PRIMARY KEY,
  "graph_node_id" uuid NOT NULL,
  "chunk_index" int NOT NULL,
  "content" text NOT NULL,
  "embedding" vector(768),
  "created_at" timestamptz NOT NULL
);

CREATE UNIQUE INDEX "chunks_unique_index" ON "chunks" ("graph_node_id", "chunk_index");

CREATE TABLE "jobs" (
  "id" uuid PRIMARY KEY,
  "kind" varchar NOT NULL,
  "queue" varchar NOT NULL,
  "payload" jsonb NOT NULL,
  "state" varchar NOT NULL,
  "priority" int DEFAULT 0,
  "attempt" int DEFAULT 0,
  "max_attempt" int DEFAULT 10,
  "worker_id" uuid,
  "scheduled_at" timestamptz,
  "attempted_at" timestamptz,
  "completed_at" timestamptz,
  "leaseexpires_at" timestamptz,
  "created_at" timestamptz,
  "updated_at" timestamptz
);

ALTER TABLE "teams" ADD FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id");

ALTER TABLE "topics" ADD FOREIGN KEY ("team_id") REFERENCES "teams" ("id");

ALTER TABLE "data_sources" ADD FOREIGN KEY ("topic_id") REFERENCES "topics" ("id");

ALTER TABLE "graph_nodes" ADD FOREIGN KEY ("data_source_id") REFERENCES "data_sources" ("id");

ALTER TABLE "graph_edges" ADD FOREIGN KEY ("root_data_source_id") REFERENCES "data_sources" ("id");

ALTER TABLE "graph_edges" ADD FOREIGN KEY ("from_node_id") REFERENCES "graph_nodes" ("id");

ALTER TABLE "graph_edges" ADD FOREIGN KEY ("to_node_id") REFERENCES "graph_nodes" ("id");

ALTER TABLE "graph_edges" ADD FOREIGN KEY ("evidence_node_id") REFERENCES "graph_nodes" ("id");

ALTER TABLE "graph_edges" ADD FOREIGN KEY ("evidence_chunk_id") REFERENCES "chunks" ("id");

ALTER TABLE "chunks" ADD FOREIGN KEY ("graph_node_id") REFERENCES "graph_nodes" ("id");
