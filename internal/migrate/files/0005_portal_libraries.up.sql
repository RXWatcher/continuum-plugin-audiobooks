CREATE TABLE IF NOT EXISTS portal_library (
  id                  BIGSERIAL PRIMARY KEY,
  name                TEXT NOT NULL,
  media_type          TEXT NOT NULL DEFAULT 'audiobook',
  backend_plugin_id   TEXT NOT NULL,
  backend_library_id  BIGINT,
  enabled             BOOL NOT NULL DEFAULT true,
  sort_order          INT NOT NULL DEFAULT 0,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS portal_library_enabled_order_idx
  ON portal_library (enabled, sort_order, id);

INSERT INTO portal_library (name, media_type, backend_plugin_id, sort_order)
SELECT 'Audiobooks', 'audiobook', target_backend_plugin_id, 0
FROM backend_config
WHERE target_backend_plugin_id <> ''
  AND NOT EXISTS (SELECT 1 FROM portal_library);
