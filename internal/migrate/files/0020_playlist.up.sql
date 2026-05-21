-- Playlists are the second half of the ABS "Collections + Playlists"
-- domain. Same shape as collections except items can reference an
-- episode in addition to a library item — this is the bit that lets
-- users build ordered queues across multiple podcasts.
--
-- The manual collection table (0002) covers static audiobook lists;
-- playlist covers ordered, podcast-friendly playback queues. The
-- distinction matches upstream ABS and keeps the data models clean.
CREATE TABLE IF NOT EXISTS playlist (
  id          TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL,
  name        TEXT NOT NULL,
  description TEXT,
  -- cover_item is the library_item_id whose cover we use as the
  -- playlist thumbnail. Optional; clients fall back to a generated
  -- collage when null.
  cover_item  TEXT,
  is_public   BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS playlist_user_idx ON playlist (user_id, LOWER(name));

CREATE TABLE IF NOT EXISTS playlist_item (
  playlist_id     TEXT NOT NULL REFERENCES playlist(id) ON DELETE CASCADE,
  library_item_id TEXT NOT NULL,
  -- episode_id is nullable: an entry referencing a podcast episode
  -- carries both library_item_id (the parent podcast) and episode_id;
  -- a book entry carries just library_item_id with episode_id = ''.
  -- We use '' rather than NULL so the (playlist, item, episode) PK
  -- works without a partial-index dance.
  episode_id      TEXT NOT NULL DEFAULT '',
  position        INT NOT NULL,
  added_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (playlist_id, library_item_id, episode_id)
);
CREATE INDEX IF NOT EXISTS playlist_item_playlist_pos_idx
  ON playlist_item (playlist_id, position);
