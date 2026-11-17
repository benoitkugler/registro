-- v0.13.0

BEGIN;
ALTER TABLE demandes ADD COLUMN Nationnalite smallint CHECK (Nationnalite IN (0, 1, 2)) NOT NULL DEFAULT 0;
ALTER TABLE demandes ALTER COLUMN Nationnalite DROP DEFAULT;
COMMIT;