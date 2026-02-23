CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    anchor_date DATE,
    prompt_day_of_week INT NOT NULL DEFAULT 0 CHECK (prompt_day_of_week BETWEEN 0 AND 6),
    prompt_hour INT NOT NULL DEFAULT 9 CHECK (prompt_hour BETWEEN 0 AND 23),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);
