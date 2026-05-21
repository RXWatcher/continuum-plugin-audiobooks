-- Smart Collections are rule-based dynamic collections — membership is
-- computed from a JSON QueryDefinition (mirroring the host's
-- internal/catalog/query_definition.go shape) rather than tracked in a
-- junction table. The manual `collection` and `collection_item` tables
-- (migration 0002) remain for static, hand-curated lists; this is a
-- separate surface.
--
-- The query_def field carries the full DSL: {match, groups[{match,
-- rules[{field, op, value}]}], sort, limit}. Audiobook-specific field
-- and sort catalogs live in Go; the DB just stores the JSON.
CREATE TABLE IF NOT EXISTS smart_collection (
  id          TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL,
  name        TEXT NOT NULL,
  description TEXT,
  color       TEXT,
  is_public   BOOLEAN NOT NULL DEFAULT FALSE,
  is_pinned   BOOLEAN NOT NULL DEFAULT FALSE,
  -- query_def is the rule DSL. Normalised + validated server-side.
  query_def   JSONB NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS smart_collection_user_idx
  ON smart_collection (user_id, is_pinned DESC, name);
CREATE INDEX IF NOT EXISTS smart_collection_public_idx
  ON smart_collection (is_public, name) WHERE is_public = TRUE;
