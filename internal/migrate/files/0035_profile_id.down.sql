-- Dropping profile_id cascades the indexes that reference it; recreate
-- the original owner-scoped indexes afterwards.
ALTER TABLE collection           DROP COLUMN IF EXISTS profile_id;
ALTER TABLE smart_collection     DROP COLUMN IF EXISTS profile_id;
ALTER TABLE playlist             DROP COLUMN IF EXISTS profile_id;
ALTER TABLE bookmark             DROP COLUMN IF EXISTS profile_id;
ALTER TABLE abs_playback_session DROP COLUMN IF EXISTS profile_id;
ALTER TABLE abs_token            DROP COLUMN IF EXISTS profile_id;

CREATE INDEX IF NOT EXISTS collection_user_pinned_idx ON collection (user_id, is_pinned DESC, name);
CREATE INDEX IF NOT EXISTS smart_collection_user_idx ON smart_collection (user_id, is_pinned DESC, name);
CREATE INDEX IF NOT EXISTS playlist_user_idx ON playlist (user_id, LOWER(name));
CREATE INDEX IF NOT EXISTS bookmark_user_book_idx ON bookmark (user_id, book_id, position_seconds);
CREATE INDEX IF NOT EXISTS abs_session_active_idx ON abs_playback_session (user_id, last_update DESC) WHERE closed_at IS NULL;
