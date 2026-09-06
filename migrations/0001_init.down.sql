-- 0001_init (down): drop everything created by 0001_init.up.sql.
-- Triggers are dropped implicitly with their tables.

DROP TABLE IF EXISTS commands;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS companies;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TYPE IF EXISTS command_status;
DROP TYPE IF EXISTS command_kind;
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS risk_level;
DROP TYPE IF EXISTS task_state;
DROP TYPE IF EXISTS agent_status;
DROP TYPE IF EXISTS agent_role;
