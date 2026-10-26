ALTER TABLE pages ADD COLUMN slug VARCHAR(200);

WITH normalized AS (
    SELECT id,
           COALESCE(NULLIF(TRIM(BOTH '-' FROM REGEXP_REPLACE(
               TRANSLATE(LOWER(name), '''’‘ʻʼ`', ''), '[^a-z]+', '-', 'g'
           )), ''), 'page') AS base_slug,
           TRANSLATE(id::TEXT, '0123456789', 'abcdefghij') AS suffix
    FROM pages
)
UPDATE pages
SET slug = RTRIM(LEFT(normalized.base_slug, 199 - LENGTH(normalized.suffix)), '-') || '-' || normalized.suffix
FROM normalized
WHERE pages.id = normalized.id;

ALTER TABLE pages
    ALTER COLUMN slug SET NOT NULL,
    ADD CONSTRAINT pages_slug_unique UNIQUE (slug),
    ADD CONSTRAINT pages_slug_format CHECK (slug ~ '^[a-z]+(-[a-z]+)*$');
