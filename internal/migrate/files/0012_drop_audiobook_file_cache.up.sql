-- Drops the audiobook_file_cache table left behind when the portal's
-- streaming cache was removed (its only Go consumer was
-- internal/streaming/cache.go, also deleted at the time). The Go store
-- helpers at internal/store/file_cache.go are removed in the same change.
DROP TABLE IF EXISTS audiobook_file_cache;
