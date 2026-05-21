-- Per-user Web Push browser subscriptions. The browser registers
-- via the Push API and POSTs the resulting {endpoint, keys}
-- payload here; the notification dispatcher sends VAPID-encrypted
-- pushes to each stored endpoint when a notification fires.
--
-- One user can have multiple subscriptions (one per browser /
-- device). Endpoint URLs are vendor-specific (fcm.googleapis.com
-- for Chrome, mozilla.com for Firefox, etc.); we don't parse them.
CREATE TABLE IF NOT EXISTS push_subscription (
  id          TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL,
  endpoint    TEXT NOT NULL,
  p256dh      TEXT NOT NULL,
  auth        TEXT NOT NULL,
  user_agent  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS push_subscription_user_idx ON push_subscription (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS push_subscription_endpoint_idx
  ON push_subscription (endpoint);
