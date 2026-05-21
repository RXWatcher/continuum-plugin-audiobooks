-- Podcasts as first-class entities alongside audiobooks. The portal hosts
-- the catalog (title/author/cover/description) and tracks each user's
-- per-episode progress; the audio bytes themselves live wherever the
-- podcast's RSS feed points (typically external CDN, not the backend
-- plugin) — `podcast_episode.audio_url` is the URL ABS clients fetch
-- directly via the play endpoint.
--
-- RSS-ingestion fields (feed_url, last_refreshed_at, refresh_interval_minutes,
-- last_error) are present from day one so the scheduled refresher can land
-- as a follow-up without another migration.

CREATE TABLE IF NOT EXISTS podcast (
  id                       TEXT PRIMARY KEY,
  library_id               BIGINT NOT NULL REFERENCES portal_library(id) ON DELETE CASCADE,
  title                    TEXT NOT NULL,
  author                   TEXT,
  description              TEXT,
  cover_url                TEXT,
  language                 TEXT,
  explicit                 BOOLEAN NOT NULL DEFAULT false,
  itunes_category          TEXT,
  feed_url                 TEXT,
  last_refreshed_at        TIMESTAMPTZ,
  refresh_interval_minutes INT NOT NULL DEFAULT 360,
  last_error               TEXT,
  created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS podcast_library_idx ON podcast (library_id);
CREATE UNIQUE INDEX IF NOT EXISTS podcast_feed_url_idx ON podcast (feed_url) WHERE feed_url IS NOT NULL;

CREATE TABLE IF NOT EXISTS podcast_episode (
  id                TEXT PRIMARY KEY,
  podcast_id        TEXT NOT NULL REFERENCES podcast(id) ON DELETE CASCADE,
  -- guid is the RSS feed's <guid> for the episode. Stable across feed
  -- refreshes; the unique key per (podcast_id, guid) prevents the
  -- refresher inserting duplicates when the feed re-emits an item.
  guid              TEXT NOT NULL,
  title             TEXT NOT NULL,
  description       TEXT,
  audio_url         TEXT NOT NULL,
  audio_mime_type   TEXT,
  audio_bytes       BIGINT,
  duration_seconds  INT NOT NULL DEFAULT 0,
  episode_index     INT,
  season_index      INT,
  published_at      TIMESTAMPTZ,
  cover_url         TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS podcast_episode_guid_idx ON podcast_episode (podcast_id, guid);
CREATE INDEX IF NOT EXISTS podcast_episode_published_idx ON podcast_episode (podcast_id, published_at DESC);

-- Per-user, per-episode progress. Mirrors the audiobook progress table's
-- shape so the existing progress-by-id endpoints can dispatch to either
-- table based on whether the id matches an audiobook book id or a
-- podcast episode id.
CREATE TABLE IF NOT EXISTS podcast_episode_progress (
  user_id          TEXT NOT NULL,
  episode_id       TEXT NOT NULL REFERENCES podcast_episode(id) ON DELETE CASCADE,
  current_seconds  INT NOT NULL DEFAULT 0,
  progress_pct     REAL NOT NULL DEFAULT 0,
  is_finished      BOOLEAN NOT NULL DEFAULT false,
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, episode_id)
);
CREATE INDEX IF NOT EXISTS podcast_episode_progress_user_updated_idx
  ON podcast_episode_progress (user_id, updated_at DESC);
