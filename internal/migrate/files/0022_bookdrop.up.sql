-- BookDrop is a watched directory the plugin scans on a schedule
-- for new audio files. Each file becomes a pending_import row with
-- whatever metadata we could parse out of the tags; an admin then
-- reviews, optionally edits, and approves — at which point the
-- plugin fires an audiobook.import event that the backend picks up.
--
-- status state machine:
--   pending  → admin hasn't touched it yet
--   editing  → admin is reviewing; metadata may be modified
--   approved → admin clicked import; queued for backend ingest
--   rejected → admin discarded; file stays on disk
--   imported → backend confirmed import; row kept for audit
CREATE TABLE IF NOT EXISTS pending_import (
  id            TEXT PRIMARY KEY,
  -- file_path is absolute on disk; the scanner records it so the
  -- admin can see what's about to be imported, and the import
  -- handler hands it to the backend.
  file_path     TEXT NOT NULL UNIQUE,
  size_bytes    BIGINT NOT NULL DEFAULT 0,
  -- Parsed metadata snapshot — title / authors / narrator / etc.
  -- Stored as JSONB so the admin can edit any field without
  -- migrations; the backend's import contract reads from here.
  metadata      JSONB NOT NULL DEFAULT '{}',
  status        TEXT NOT NULL DEFAULT 'pending',
  -- error_message captures backend rejection so the admin can see
  -- why an import failed and re-try after editing.
  error_message TEXT,
  -- target_library_id is the admin's pick for which portal library
  -- the book should land in. Null until admin sets it.
  target_library_id BIGINT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS pending_import_status_idx
  ON pending_import (status, created_at DESC);
