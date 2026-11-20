-- v0.13.0
-- SecuriteSociale est désormais porté par Personne (et non plus par Ficheequipier)
BEGIN;

ALTER TABLE personnes
    ADD COLUMN SecuriteSociale text NOT NULL DEFAULT '';

UPDATE
    personnes
SET
    SecuriteSociale = f.SecuriteSociale
FROM
    ficheequipiers f
WHERE
    f.IdPersonne = personnes.Id;

ALTER TABLE personnes
    ALTER COLUMN SecuriteSociale DROP DEFAULT;

ALTER TABLE ficheequipiers
    DROP COLUMN SecuriteSociale;

COMMIT;
