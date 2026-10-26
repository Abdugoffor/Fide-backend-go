ALTER TABLE applications_data ADD COLUMN name VARCHAR(200) NOT NULL DEFAULT '';

UPDATE applications_data AS data
SET name = field.name
FROM fields AS field
WHERE data.field = field.id;

ALTER TABLE applications_data ALTER COLUMN name DROP DEFAULT;
