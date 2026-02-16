-- ajoute les champs equipiers (repere) sur une nouvelle table
BEGIN;
--
CREATE TABLE ficheequipiers (
    IdPersonne integer NOT NULL,
    SecuriteSociale text NOT NULL,
    Fonctionnaire boolean NOT NULL,
    Diplome smallint CHECK (Diplome IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)) NOT NULL,
    Approfondissement smallint CHECK (Approfondissement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Profession text NOT NULL,
    EtatCivil smallint CHECK (EtatCivil IN (0, 1, 2)) NOT NULL,
    NombreEnfants integer NOT NULL,
    ExperienceTravailJeunes text NOT NULL,
    ParcoursSpirituel text NOT NULL,
    Eglise text NOT NULL,
    Recommandation jsonb NOT NULL,
    Sante text NOT NULL,
    AssuranceMaladie text NOT NULL,
    AssuranceAccident text NOT NULL,
    MembreAssoPermanent boolean NOT NULL
);
INSERT INTO ficheequipiers
SELECT
    Id,
    '',
    Fonctionnaire,
    Diplome,
    Approfondissement,
    Profession,
    0,
    0,
    '',
    '',
    '',
    '{"Nom": "", "Prenom": "", "Mail":"", "Tel": ""}',
    '',
    '',
    '',
    FALSE
FROM
    personnes;
--
ALTER TABLE personnes
    DROP COLUMN NomJeuneFille;
ALTER TABLE personnes
    DROP COLUMN Profession;
ALTER TABLE personnes
    DROP COLUMN Etudiant;
ALTER TABLE personnes
    DROP COLUMN Fonctionnaire;
ALTER TABLE personnes
    DROP COLUMN Diplome;
ALTER TABLE personnes
    DROP COLUMN Approfondissement;
ROLLBACK;

