-- Phase 2 per-profile re-keying. profile_id '' is the canonical primary
-- profile; existing rows backfill to '' so every current user keeps
-- their data under their primary profile. The progress table is
-- re-keyed by a separate later migration (its primary key changes).

ALTER TABLE collection           ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';
ALTER TABLE smart_collection     ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';
ALTER TABLE playlist             ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';
ALTER TABLE bookmark             ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';
ALTER TABLE abs_playback_session ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';
ALTER TABLE abs_token            ADD COLUMN IF NOT EXISTS profile_id TEXT NOT NULL DEFAULT '';

-- Extend the owner-scoped indexes with profile_id.
DROP INDEX IF EXISTS collection_user_pinned_idx;
CREATE INDEX collection_user_pinned_idx ON collection (user_id, profile_id, is_pinned DESC, name);
DROP INDEX IF EXISTS smart_collection_user_idx;
CREATE INDEX smart_collection_user_idx ON smart_collection (user_id, profile_id, is_pinned DESC, name);
DROP INDEX IF EXISTS playlist_user_idx;
CREATE INDEX playlist_user_idx ON playlist (user_id, profile_id, LOWER(name));
DROP INDEX IF EXISTS bookmark_user_book_idx;
CREATE INDEX bookmark_user_book_idx ON bookmark (user_id, profile_id, book_id, position_seconds);
DROP INDEX IF EXISTS abs_session_active_idx;
CREATE INDEX abs_session_active_idx ON abs_playback_session (user_id, profile_id, last_update DESC) WHERE closed_at IS NULL;
