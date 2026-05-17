ALTER TABLE backend_config
  DROP COLUMN IF EXISTS target_request_provider_installation_id,
  DROP COLUMN IF EXISTS target_request_provider_plugin_id;
