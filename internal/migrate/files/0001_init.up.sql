CREATE TABLE IF NOT EXISTS backend_config (
  id                              INT PRIMARY KEY DEFAULT 1,
  target_backend_plugin_id        TEXT NOT NULL DEFAULT '',
  auto_approve_requests           BOOL NOT NULL DEFAULT false,
  streaming_mode                  TEXT NOT NULL DEFAULT 'proxy',
  cache_dir                       TEXT,
  cache_max_size_gb               INT NOT NULL DEFAULT 50,
  cache_download_concurrency      INT NOT NULL DEFAULT 2,
  path_remappings                 JSONB NOT NULL DEFAULT '[]'::jsonb,
  abs_jwt_secret                  BYTEA NOT NULL,
  abs_access_token_ttl_hours      INT NOT NULL DEFAULT 24,
  abs_refresh_token_ttl_days      INT NOT NULL DEFAULT 30,
  updated_at                      TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT backend_config_singleton CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS progress (
  user_id          TEXT NOT NULL,
  book_id          TEXT NOT NULL,
  current_seconds  INT NOT NULL DEFAULT 0,
  progress_pct     REAL NOT NULL DEFAULT 0,
  is_finished      BOOL NOT NULL DEFAULT false,
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, book_id)
);
CREATE INDEX IF NOT EXISTS progress_user_updated_idx ON progress (user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS bookmark (
  id                TEXT PRIMARY KEY,
  user_id           TEXT NOT NULL,
  book_id           TEXT NOT NULL,
  position_seconds  INT NOT NULL,
  chapter_id        TEXT,
  note              TEXT NOT NULL DEFAULT '',
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS bookmark_user_book_idx ON bookmark (user_id, book_id, position_seconds);

CREATE TABLE IF NOT EXISTS rating (
  user_id     TEXT NOT NULL,
  book_id     TEXT NOT NULL,
  rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, book_id)
);

CREATE TABLE IF NOT EXISTS request (
  id                  TEXT PRIMARY KEY,
  user_id             TEXT NOT NULL,
  title               TEXT NOT NULL,
  author              TEXT,
  isbn                TEXT,
  status              TEXT NOT NULL DEFAULT 'pending',
  target_plugin_id    TEXT NOT NULL DEFAULT '',
  external_id         TEXT,
  denied_reason       TEXT,
  failure_reason      TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  fulfilled_at        TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS request_status_created_idx ON request (status, created_at DESC);
CREATE INDEX IF NOT EXISTS request_user_created_idx ON request (user_id, created_at DESC);
