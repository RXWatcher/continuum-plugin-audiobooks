-- Discrete listening-session events. progress.updated_at already
-- gives us "did the user touch this book on date X?" but it doesn't
-- capture session length or per-session boundaries — the heatmap +
-- year-in-review surfaces need that granularity.
--
-- Sessions are recorded by the SPA on close (`/me/reading-sessions`
-- POST) and aggregated read-side. Open-ended sessions (started_at
-- but no ended_at) are tolerated for resumability — a session
-- without ended_at is in-flight; the close call sets ended_at +
-- seconds_played.
CREATE TABLE IF NOT EXISTS reading_session (
  id              TEXT PRIMARY KEY,
  user_id         TEXT NOT NULL,
  book_id         TEXT NOT NULL,
  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at        TIMESTAMPTZ,
  seconds_played  INT NOT NULL DEFAULT 0,
  -- device_label is an optional free-form tag the client supplies
  -- (e.g. "iPhone", "Web — Firefox") so the stats surface can
  -- group by device. Nullable.
  device_label    TEXT
);
CREATE INDEX IF NOT EXISTS reading_session_user_started_idx
  ON reading_session (user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS reading_session_user_book_idx
  ON reading_session (user_id, book_id, started_at DESC);
