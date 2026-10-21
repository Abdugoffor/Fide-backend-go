CREATE TABLE IF NOT EXISTS fields (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES pages(id),
    sort INTEGER NOT NULL CHECK (sort > 0),
    type VARCHAR(20) NOT NULL CHECK (type IN ('text', 'textarea', 'email', 'tel', 'number', 'date', 'select', 'radio', 'checkbox', 'file')),
    options JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(options) = 'array'),
    name VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    validation JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(validation) = 'object'),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fields_page_sort_unique UNIQUE (page_id, sort),
    CONSTRAINT fields_page_name_unique UNIQUE (page_id, name)
);
