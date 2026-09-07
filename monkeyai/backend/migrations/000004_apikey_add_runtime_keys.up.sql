BEGIN;

CREATE TABLE api_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id),
    name text NOT NULL,
    key_prefix text NOT NULL,
    key_hash text NOT NULL UNIQUE,
    scopes text[] NOT NULL,
    expires_at timestamptz NOT NULL,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CONSTRAINT api_keys_name_check CHECK (btrim(name) <> ''),
    CONSTRAINT api_keys_scopes_check CHECK (
        cardinality(scopes) > 0
        AND scopes <@ ARRAY['model:invoke', 'mcp:invoke']::text[]
    ),
    CONSTRAINT api_keys_expiry_check CHECK (expires_at > created_at)
);

CREATE INDEX api_keys_user_active_idx
    ON api_keys (user_id, created_at DESC)
    WHERE revoked_at IS NULL;

CREATE INDEX api_keys_expiry_idx
    ON api_keys (expires_at)
    WHERE revoked_at IS NULL;

COMMIT;
