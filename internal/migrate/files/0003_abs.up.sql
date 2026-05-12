CREATE TABLE IF NOT EXISTS abs_token (
  id           TEXT PRIMARY KEY,
  user_id      TEXT NOT NULL,
  jti          TEXT UNIQUE NOT NULL,
  device_id    TEXT,
  device_name  TEXT,
  device_info  JSONB,
  last_used_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at   TIMESTAMPTZ NOT NULL,
  revoked_at   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS abs_playback_session (
  id            TEXT PRIMARY KEY,
  user_id       TEXT NOT NULL,
  book_id       TEXT NOT NULL,
  device_id     TEXT NOT NULL,
  device_info   JSONB,
  play_method   TEXT NOT NULL DEFAULT 'directplay',
  media_player  TEXT,
  start_time    INT NOT NULL DEFAULT 0,
  current_time_ms INT NOT NULL DEFAULT 0,
  started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_update   TIMESTAMPTZ NOT NULL DEFAULT now(),
  closed_at     TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS abs_session_active_idx ON abs_playback_session (user_id, last_update DESC) WHERE closed_at IS NULL;
