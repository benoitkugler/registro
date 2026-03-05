-- v0.10.2
-- add more context on event validation

BEGIN;
ALTER TABLE event_validations
    DROP COLUMN guard;
--
--

ALTER TABLE event_validations
    ADD COLUMN IsBackoffice boolean;
ALTER TABLE event_validations
    ADD COLUMN guard smallint;
--
-- Add values
-- validation from backoffice are marked by a NULL camp

UPDATE
    event_validations
SET
    IsBackoffice = (IdCamp IS NULL);
-- choose a random camp (manual inspection is required)
UPDATE
    event_validations
SET
    IdCamp = (
        SELECT
            id
        FROM
            camps
        LIMIT 1)
WHERE
    IsBackoffice;
-- guard value
UPDATE
    event_validations
SET
    guard = 1;
--
-- Setup constraints again

ALTER TABLE event_validations
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;
ALTER TABLE event_validations
    ALTER COLUMN IdCamp SET NOT NULL;
ALTER TABLE event_validations
    ALTER COLUMN IsBackoffice SET NOT NULL;
ALTER TABLE event_validations
    ALTER COLUMN guard SET NOT NULL;
ALTER TABLE event_validations
    ALTER COLUMN guard SET DEFAULT 1
    /* EventKind.Validation */
;
ALTER TABLE event_validations
    ADD CHECK (guard = 1
    /* EventKind.Validation */);
COMMIT;

