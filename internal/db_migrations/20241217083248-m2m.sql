
-- +migrate Up
CREATE TABLE m2m_sessions (
    id text PRIMARY KEY,
    subject_id text NOT NULL,
    refresh_token text NOT NULL, -- hashed    
    roles text[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_authenticated_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    service_id uuid NOT NULL REFERENCES services(id)
);

ALTER TABLE m2m_sessions 
ADD CONSTRAINT roles_not_empty 
CHECK (array_length(roles, 1) > 0);

CREATE INDEX idx_m2m_sessions_subject_id ON m2m_sessions(subject_id);

CREATE TABLE m2m_session_templates (
    id text NOT NULL,
    roles text[],
    service_id uuid NOT NULL REFERENCES services(id),

    PRIMARY KEY (id, service_id)
);

ALTER TABLE m2m_session_templates 
ADD CONSTRAINT roles_not_empty 
CHECK (array_length(roles, 1) > 0);

-- +migrate Down
DROP INDEX IF EXISTS idx_m2m_sessions_subject_id;
DROP TABLE IF EXISTS m2m_sessions;
DROP TABLE IF EXISTS m2m_session_templates;
