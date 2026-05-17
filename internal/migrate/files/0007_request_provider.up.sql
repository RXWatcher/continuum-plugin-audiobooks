ALTER TABLE backend_config
  ADD COLUMN IF NOT EXISTS target_request_provider_plugin_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS target_request_provider_installation_id TEXT NOT NULL DEFAULT '';

UPDATE backend_config
SET target_request_provider_plugin_id = target_backend_plugin_id
WHERE target_request_provider_plugin_id = '';

UPDATE backend_config
SET target_request_provider_installation_id = target_backend_installation_id
WHERE target_request_provider_installation_id = ''
  AND target_backend_installation_id <> '';
