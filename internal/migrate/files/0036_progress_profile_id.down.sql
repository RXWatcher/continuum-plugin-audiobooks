ALTER TABLE progress DROP CONSTRAINT progress_pkey;
ALTER TABLE progress ADD PRIMARY KEY (user_id, book_id);
ALTER TABLE progress DROP COLUMN IF EXISTS profile_id;
CREATE INDEX IF NOT EXISTS progress_user_updated_idx ON progress (user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS progress_user_updated_visible_idx ON progress (user_id, updated_at DESC) WHERE hidden_from_continue = FALSE;
