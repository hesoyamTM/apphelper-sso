CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(20) NOT NULL,
    surname VARCHAR(20) NOT NULL,
    email VARCHAR(20) NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE
)