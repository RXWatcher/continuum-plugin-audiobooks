CREATE TABLE IF NOT EXISTS audiobook_file_cache (
  id                TEXT PRIMARY KEY,
  cache_key         TEXT UNIQUE NOT NULL,
  book_id           TEXT NOT NULL,
  file_idx          INT,
  filename          TEXT NOT NULL,
  mime_type         TEXT NOT NULL,
  content_length    BIGINT NOT NULL,
  codec             TEXT,
  duration_seconds  INT,
  status            TEXT NOT NULL DEFAULT 'pending',
  download_progress REAL NOT NULL DEFAULT 0,
  error_message     TEXT,
  relative_path     TEXT NOT NULL,
  bytes_on_disk     BIGINT NOT NULL DEFAULT 0,
  last_accessed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS file_cache_status_accessed_idx ON audiobook_file_cache (status, last_accessed_at);
CREATE INDEX IF NOT EXISTS file_cache_book_file_idx ON audiobook_file_cache (book_id, file_idx);
