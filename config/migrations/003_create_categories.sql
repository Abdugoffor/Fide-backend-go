CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    parent_id BIGINT,
    name VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_categories_tournament_id ON categories(tournament_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id, id DESC);
