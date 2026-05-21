-- Bookmark change log for replica sync. Every bookmark mutation
-- (upsert / delete) appends a row here with the originating HLC
-- timestamp; replicas pull changes since their last seen hlc.
--
-- Row-level LWW with tombstones: the highest hlc for a given
-- bookmark_id wins; deletes write op='delete' rather than
-- removing earlier rows so peers can observe the deletion.
CREATE TABLE IF NOT EXISTS bookmark_change (
  hlc          TEXT PRIMARY KEY,
  user_id      TEXT NOT NULL,
  bookmark_id  TEXT NOT NULL,
  op           TEXT NOT NULL,    -- 'upsert' | 'delete'
  payload      JSONB NOT NULL DEFAULT '{}',
  origin_node  TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS bookmark_change_user_hlc_idx
  ON bookmark_change (user_id, hlc);
