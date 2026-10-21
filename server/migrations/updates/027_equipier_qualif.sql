-- v0.12.3
-- Remplace l'enum Approfondissement par un champ Qualifications
BEGIN;

ALTER TABLE
    ficheequipiers RENAME COLUMN Approfondissement TO Qualifications;

CREATE TYPE Qualifications AS (
    SB boolean,
    Canoe boolean,
    Voile boolean
);

ALTER TABLE
    ficheequipiers DROP CONSTRAINT ficheequipiers_approfondissement_check;

ALTER TABLE
    ficheequipiers
ALTER COLUMN
    Qualifications TYPE Qualifications USING '(f,f,f)' :: Qualifications;

ALTER TABLE
    ficheequipiers
ALTER COLUMN
    Qualifications
SET
    NOT NULL;

COMMIT;