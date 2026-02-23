CREATE TABLE portraits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_path TEXT NOT NULL,
    portrait_year INT NOT NULL,
    is_manual_upload BOOLEAN NOT NULL DEFAULT FALSE,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_portraits_user_id ON portraits (user_id);
CREATE INDEX idx_portraits_user_year ON portraits (user_id, portrait_year);
