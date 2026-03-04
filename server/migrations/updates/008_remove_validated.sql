-- v0.10.1
-- remove IsValidated field from dossiers

BEGIN;
ALTER TABLE dossiers
    DROP COLUMN IsValidated;
COMMIT;

