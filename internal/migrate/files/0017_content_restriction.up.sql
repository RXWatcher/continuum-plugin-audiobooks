-- Per-user content restrictions for family / child accounts. Admin
-- writes one row per restricted user; the catalog + personalized +
-- search handlers drop matching items before they leave the plugin.
-- Mirrors grimmory's ContentRestriction model but uses Postgres
-- arrays directly rather than a per-rule junction table — at our
-- scale the array approach reads faster and is simpler to admin.
CREATE TABLE IF NOT EXISTS content_restriction (
  user_id            TEXT PRIMARY KEY,
  -- Each list is "blocked when item has any of these". Empty = no
  -- filter on that dimension. Genres / tags / authors / narrators
  -- match case-insensitively at evaluate time.
  blocked_genres     TEXT[] NOT NULL DEFAULT '{}',
  blocked_tags       TEXT[] NOT NULL DEFAULT '{}',
  blocked_authors    TEXT[] NOT NULL DEFAULT '{}',
  blocked_narrators  TEXT[] NOT NULL DEFAULT '{}',
  blocked_libraries  BIGINT[] NOT NULL DEFAULT '{}',
  -- explicit_blocked filters items marked explicit upstream (the
  -- audiobook backend's `explicit` field on detail). Useful for
  -- "child mode" without enumerating individual tags.
  explicit_blocked   BOOLEAN NOT NULL DEFAULT FALSE,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
