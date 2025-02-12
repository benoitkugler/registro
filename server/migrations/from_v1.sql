BEGIN;
CREATE TABLE personnes2 (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    Prenom text NOT NULL,
    Sexe smallint CHECK (Sexe IN (0, 1, 2)) NOT NULL,
    DateNaissance date NOT NULL,
    VilleNaissance text NOT NULL,
    DepartementNaissance text NOT NULL,
    Nationnalite smallint CHECK (Nationnalite IN (0, 1, 2)) NOT NULL,
    Tels text[],
    Mail text NOT NULL,
    Adresse text NOT NULL,
    CodePostal text NOT NULL,
    Ville text NOT NULL,
    Pays text NOT NULL,
    SecuriteSociale text NOT NULL,
    NomJeuneFille text NOT NULL,
    Profession text NOT NULL,
    Etudiant boolean NOT NULL,
    Fonctionnaire boolean NOT NULL,
    Diplome smallint CHECK (Diplome IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)) NOT NULL,
    Approfondissement smallint CHECK (Approfondissement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Publicite jsonb NOT NULL,
    IsTemp boolean NOT NULL
);
INSERT INTO personnes2
SELECT
    id,
    nom,
    prenom,
    (
        CASE WHEN Sexe = 'F' THEN
            1
        WHEN Sexe = 'M' THEN
            2
        ELSE
            0
        END),
    date_naissance,
    ville_naissance,
    departement_naissance,
    0, -- Nationnalite
    Tels,
    Mail,
    Adresse,
    code_postal,
    Ville,
    Pays,
    securite_sociale,
    nom_jeune_fille,
    Profession,
    Etudiant,
    Fonctionnaire,
    0, -- Diplome,
    0, -- Approfondissement,
    '{"VersionPapier": false,"PubHiver": false,"PubEte": false,"EchoRocher": false,"Eonews": false}',
    is_temporaire
FROM
    personnes;
ALTER TABLE personnes RENAME TO personnes_old;
ALTER TABLE personnes2 RENAME TO personnes;
COMMIT;

