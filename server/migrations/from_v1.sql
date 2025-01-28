BEGIN;
CREATE TABLE personnes2 (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    NomJeuneFille text NOT NULL,
    Prenom text NOT NULL,
    DateNaissance date NOT NULL,
    VilleNaissance text NOT NULL,
    DepartementNaissance text NOT NULL,
    Sexe smallint CHECK (Sexe IN (0, 1, 2)) NOT NULL,
    Tels text[],
    Mail text NOT NULL,
    Adresse text NOT NULL,
    CodePostal text NOT NULL,
    Ville text NOT NULL,
    Pays text NOT NULL,
    SecuriteSociale text NOT NULL,
    Profession text NOT NULL,
    Etudiant boolean NOT NULL,
    Fonctionnaire boolean NOT NULL,
    Diplome smallint CHECK (Diplome IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)) NOT NULL,
    Approfondissement smallint CHECK (Approfondissement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    IsTemp boolean NOT NULL
);
INSERT INTO personnes2
SELECT
    Id,
    Nom,
    nom_jeune_fille,
    Prenom,
    date_naissance,
    ville_naissance,
    departement_naissance,
    (
        CASE WHEN Sexe = 'F' THEN
            1
        WHEN Sexe = 'M' THEN
            2
        ELSE
            0
        END),
    Tels,
    Mail,
    Adresse,
    code_postal,
    Ville,
    Pays,
    securite_sociale,
    Profession,
    Etudiant,
    Fonctionnaire,
    0, -- Diplome,
    0, -- Approfondissement,
    is_temporaire
FROM
    personnes;
ALTER TABLE personnes RENAME TO personnes_old;
ALTER TABLE personnes2 RENAME TO personnes;
COMMIT;

-- ALTER TABLE personnes
--     DROP COLUMN fiche_sanitaire;
-- ALTER TABLE personnes
--     DROP COLUMN eonews;
-- ALTER TABLE personnes
--     DROP COLUMN cotisation;
-- ALTER TABLE personnes
--     DROP COLUMN quotient_familial;
-- ALTER TABLE personnes
--     DROP COLUMN rang_membre_asso;
-- ALTER TABLE personnes
--     DROP COLUMN version_papier;
-- ALTER TABLE personnes
--     DROP COLUMN pub_hiver;
-- ALTER TABLE personnes
--     DROP COLUMN pub_ete;
-- ALTER TABLE personnes
--     DROP COLUMN echo_rocher;
