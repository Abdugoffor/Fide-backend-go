CREATE UNIQUE INDEX IF NOT EXISTS categories_id_tournament_unique ON categories(id, tournament_id);

CREATE TABLE IF NOT EXISTS pages (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    sort INTEGER NOT NULL CHECK (sort >= 3),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pages_tournament_fk FOREIGN KEY (tournament_id) REFERENCES tournaments(id),
    CONSTRAINT pages_category_tournament_fk FOREIGN KEY (category_id, tournament_id) REFERENCES categories(id, tournament_id),
    CONSTRAINT pages_category_sort_unique UNIQUE (category_id, sort)
);

CREATE INDEX IF NOT EXISTS idx_pages_tournament_category_sort ON pages(tournament_id, category_id, sort);
