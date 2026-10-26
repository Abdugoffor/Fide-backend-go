CREATE TABLE IF NOT EXISTS applications_data (
    id BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL,
    page BIGINT NOT NULL,
    slug VARCHAR(200) NOT NULL,
    field BIGINT NOT NULL,
    value TEXT NOT NULL CHECK (LENGTH(TRIM(value)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_applications_data_application_page ON applications_data(application_id, page);
