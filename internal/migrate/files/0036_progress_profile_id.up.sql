-- Phase 2: re-key progress by profile. profile_id joins the primary key
-- ('' = primary profile). Done in its own migration because the PK swap
-- must land together with the progress store-code change.
ALTER TABLE progress ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';
ALTER TABLE progress DROP CONSTRAINT progress_pkey;
ALTER TABLE progress ADD PRIMARY KEY (user_id, profile_id, book_id);
DROP INDEX IF EXISTS progress_user_updated_idx;
CREATE INDEX progress_user_updated_idx ON progress (user_id, profile_id, updated_at DESC);
DROP INDEX IF EXISTS progress_user_updated_visible_idx;
CREATE INDEX progress_user_updated_visible_idx ON progress (user_id, profile_id, updated_at DESC) WHERE hidden_from_continue = FALSE;
