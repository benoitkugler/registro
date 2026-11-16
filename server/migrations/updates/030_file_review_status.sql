-- Store the director's review state for each uploaded file.

BEGIN;
ALTER TABLE files
    ADD COLUMN ReviewStatus smallint NOT NULL DEFAULT 0
        CHECK (ReviewStatus IN (0, 1, 2));
ALTER TABLE fichesanitaires
    ADD COLUMN ReviewStatus smallint NOT NULL DEFAULT 0
        CHECK (ReviewStatus IN (0, 1, 2));
COMMIT;
