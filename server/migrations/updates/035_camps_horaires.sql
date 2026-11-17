-- v0.13.0

BEGIN;
CREATE TYPE Horaire AS (
    Valid boolean,
    Heure smallint,
    Minute smallint
);
ALTER TABLE camps ADD COLUMN HoraireDebut Horaire NOT NULL DEFAULT ('f', 0, 0);
ALTER TABLE camps ADD COLUMN HoraireFin Horaire NOT NULL DEFAULT ('f', 0, 0);
ALTER TABLE camps ALTER COLUMN HoraireDebut DROP DEFAULT;
ALTER TABLE camps ALTER COLUMN HoraireFin DROP DEFAULT;
COMMIT;