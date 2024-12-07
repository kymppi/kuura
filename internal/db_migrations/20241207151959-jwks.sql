
-- +migrate Up
CREATE TABLE jwk_private (
    id text PRIMARY KEY,
    service_id uuid NOT NULL,
    encrypted_key_data bytea NOT NULL,
    nonce bytea NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE jwk_public_keys (
    id text PRIMARY KEY REFERENCES jwk_private(id) ON DELETE CASCADE,
    service_id uuid NOT NULL,
    key_data bytea NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE services (
    id uuid PRIMARY KEY, -- uuidv7
    jwt_audience text NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    modified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    description text
);

CREATE TABLE service_key_states (
    service_id uuid NOT NULL,
    jwk_private_id text NOT NULL,
    status text NOT NULL CHECK (status IN ('future', 'current', 'retired')),
    PRIMARY KEY (service_id, jwk_private_id),
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    FOREIGN KEY (jwk_private_id) REFERENCES jwk_private(id) ON DELETE RESTRICT
);

CREATE INDEX idx_jwk_private_service_id ON jwk_private(service_id);
CREATE INDEX idx_jwk_public_keys_service_id ON jwk_public_keys(service_id);
CREATE INDEX idx_service_key_states_service_id ON service_key_states(service_id);
CREATE INDEX idx_service_key_states_jwk_private_id ON service_key_states(jwk_private_id);

-- +migrate Down
DROP TABLE IF EXISTS service_key_states;
DROP TABLE IF EXISTS jwk_public_keys;
DROP TABLE IF EXISTS jwk_private;
DROP TABLE IF EXISTS services;
