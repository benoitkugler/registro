-- v0.12.3
-- remove equivalent bafa & bafd
BEGIN;

-- This migration assumes there is actually no 
-- file with catégorie equivalent bafa or equivalent bafd
-- we want to preserve the id = categorie implicit invariant (might be useful in the future)...
-- so we have to adjust the related tables
UPDATE
    demande_equipiers
SET
    IdDemande = IdDemande - 2
WHERE
    IdDemande >= 12
    AND IdDemande <= 15;

UPDATE
    demande_camps
SET
    IdDemande = IdDemande - 2
WHERE
    IdDemande >= 12
    AND IdDemande <= 15;

UPDATE
    file_personnes
SET
    IdDemande = IdDemande - 2
WHERE
    IdDemande >= 12
    AND IdDemande <= 15;

-- remove last demandes, unused right now
DELETE FROM
    demandes
WHERE
    Categorie >= 14
    AND Categorie <= 15;

-- adjust constraint check
ALTER TABLE
    demandes DROP CONSTRAINT demandes_categorie_check;

ALTER TABLE
    demandes
ADD
    CONSTRAINT demandes_categorie_check CHECK (
        Categorie IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13)
    );

-- enforce the rule
ALTER TABLE
    demandes
ADD
    CONSTRAINT constraint_categorie_is_id CHECK(
        Categorie = 0
        OR Categorie = Id
    );

COMMIT;