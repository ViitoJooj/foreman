-- 0001_init: base schema for the autodev harness.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enum types -----------------------------------------------------------------

CREATE TYPE agent_role AS ENUM ('task_creator', 'coder', 'tester', 'pr_reviewer');
CREATE TYPE agent_status AS ENUM ('idle', 'working', 'paused');
CREATE TYPE task_state AS ENUM (
    'created', 'researching', 'queued', 'coding', 'testing', 'reviewing',
    'merged', 'needs_human', 'rejected', 'failed_build', 'failed_test'
);
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high');
CREATE TYPE message_type AS ENUM ('request', 'response', 'status');
CREATE TYPE command_kind AS ENUM ('kill', 'panic', 'pause', 'resume');
CREATE TYPE command_status AS ENUM ('pending', 'applied', 'failed');

-- updated_at trigger -------------------------------------------------------

CREATE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- companies --------------------------------------------------------------

CREATE TABLE companies (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    slug       text NOT NULL UNIQUE,
    repo_owner text NOT NULL,
    repo_name  text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (repo_owner, repo_name)
);

CREATE TRIGGER companies_set_updated_at BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- agents --------------------------------------------------------------

CREATE TABLE agents (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name       text NOT NULL,
    role       agent_role NOT NULL,
    status     agent_status NOT NULL DEFAULT 'idle',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (company_id, name)
);

CREATE INDEX idx_agents_company_id ON agents (company_id);

CREATE TRIGGER agents_set_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- channels --------------------------------------------------------------

CREATE TABLE channels (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (company_id, name)
);

CREATE INDEX idx_channels_company_id ON channels (company_id);

-- tasks --------------------------------------------------------------

CREATE TABLE tasks (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    title       text NOT NULL,
    description text NOT NULL DEFAULT '',
    state       task_state NOT NULL DEFAULT 'created',
    risk        risk_level NOT NULL,
    branch      text NOT NULL DEFAULT '',
    pr_number   integer NOT NULL DEFAULT 0,
    retries     integer NOT NULL DEFAULT 0,
    max_retries integer NOT NULL DEFAULT 3,
    assignee_id uuid REFERENCES agents (id) ON DELETE SET NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_company_id ON tasks (company_id);
CREATE INDEX idx_tasks_state ON tasks (state);
CREATE INDEX idx_tasks_assignee_id ON tasks (assignee_id);

CREATE TRIGGER tasks_set_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- messages --------------------------------------------------------------

CREATE TABLE messages (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id        uuid NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    type              message_type NOT NULL,
    from_agent_id     uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    to_agent_id       uuid REFERENCES agents (id) ON DELETE SET NULL,
    in_reply_to       uuid REFERENCES messages (id) ON DELETE SET NULL,
    requires_response boolean NOT NULL DEFAULT false,
    body              text NOT NULL DEFAULT '',
    payload           jsonb NOT NULL DEFAULT '{}',
    created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_channel_created ON messages (channel_id, created_at);
CREATE INDEX idx_messages_from_agent_id ON messages (from_agent_id);

-- commands --------------------------------------------------------------

CREATE TABLE commands (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      uuid REFERENCES companies (id) ON DELETE CASCADE,
    target_agent_id uuid REFERENCES agents (id) ON DELETE CASCADE,
    kind            command_kind NOT NULL,
    status          command_status NOT NULL DEFAULT 'pending',
    reason          text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    applied_at      timestamptz
);

CREATE INDEX idx_commands_status ON commands (status);
CREATE INDEX idx_commands_company_id ON commands (company_id);

-- Row level security: deny-all, no policies. The Go runtime connects over a
-- direct Postgres role that bypasses RLS; frontend-facing policies are added
-- in a later migration when the dashboard talks to Supabase directly.

ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE commands ENABLE ROW LEVEL SECURITY;
