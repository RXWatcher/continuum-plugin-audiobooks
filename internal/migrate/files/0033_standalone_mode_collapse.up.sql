-- Collapse the three-mode standalone_login_mode to enabled/disabled.
UPDATE backend_config
   SET standalone_login_mode = 'enabled'
 WHERE standalone_login_mode IN ('opt_in', 'all_accounts');
