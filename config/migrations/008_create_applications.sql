CREATE TABLE IF NOT EXISTS applications (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    email VARCHAR(254) NOT NULL,
    birth_date DATE NOT NULL,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT applications_tournament_fk FOREIGN KEY (tournament_id) REFERENCES tournaments(id),
    CONSTRAINT applications_category_tournament_fk FOREIGN KEY (category_id, tournament_id) REFERENCES categories(id, tournament_id)
);

CREATE INDEX IF NOT EXISTS idx_applications_tournament_category_id ON applications(tournament_id, category_id, id DESC);
