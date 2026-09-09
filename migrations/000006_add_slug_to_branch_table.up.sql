-- Step 1: Add slug column as nullable first
ALTER TABLE branch ADD COLUMN slug VARCHAR(255) NULL AFTER name;

-- Step 2: Populate slug for existing rows (using branch name or id fallback)
UPDATE branch SET slug = LOWER(REPLACE(name, ' ', '-')) WHERE slug IS NULL OR slug = '';

-- Step 3: Make slug NOT NULL and UNIQUE
ALTER TABLE branch MODIFY COLUMN slug VARCHAR(255) NOT NULL UNIQUE;
