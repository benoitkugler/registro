-- v0.12.1
BEGIN;

ALTER TABLE
    camps
ADD
    COLUMN PlacesReservees integer;

UPDATE
    camps
SET
    PlacesReservees = 0;

ALTER TABLE
    camps
ALTER COLUMN
    PlacesReservees
SET
    NOT NULL;

COMMIT;