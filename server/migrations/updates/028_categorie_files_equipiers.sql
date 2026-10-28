-- v0.12.4
-- remove Haccp & CertMedicalCuisine
BEGIN;

-- This migration assumes there is actually no 
-- file with catégorie Haccp & CertMedicalCuisine (value 9 & 10)
-- we want to preserve the id = categorie implicit invariant (might be useful in the future)...
-- so we have to adjust the related tables
UPDATE
    demande_equipiers
SET
    IdDemande = IdDemande - 2
WHERE
    IdDemande >= 11
    AND IdDemande <= 13;

UPDATE
    demande_camps
SET
    IdDemande = IdDemande - 2
WHERE
    IdDemande >= 11
    AND IdDemande <= 13;

;

UPDATE
    file_personnes
SET
    IdDemande = IdDemande - 2
WHERE
    IdDemande >= 11
    AND IdDemande <= 13;

-- remove last demandes, unused right now
DELETE FROM
    demandes
WHERE
    Categorie >= 12
    AND Categorie <= 13;

-- adjust constraint check
ALTER TABLE
    demandes DROP CONSTRAINT demandes_categorie_check;

ALTER TABLE
    demandes
ADD
    CONSTRAINT demandes_categorie_check CHECK (
        Categorie IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
    );

COMMIT;