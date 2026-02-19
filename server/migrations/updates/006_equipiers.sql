-- ajoute les champs equipiers (repere) sur une nouvelle table
BEGIN;
--
CREATE TABLE ficheequipiers (
    IdPersonne integer NOT NULL,
    SecuriteSociale text NOT NULL,
    Fonctionnaire boolean NOT NULL,
    Diplome smallint CHECK (Diplome IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)) NOT NULL,
    Approfondissement smallint CHECK (Approfondissement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    EtatCivil smallint CHECK (EtatCivil IN (0, 1, 2)) NOT NULL,
    NombreEnfants integer NOT NULL,
    Formation text NOT NULL,
    Profession text NOT NULL,
    ExperienceTravailJeunes text NOT NULL,
    ParcoursSpirituel text NOT NULL,
    Eglise text NOT NULL,
    Recommandation jsonb NOT NULL,
    Sante text NOT NULL,
    AssuranceMaladie text NOT NULL,
    AssuranceAccident text NOT NULL,
    DemandeMembreAssoPermanent boolean NOT NULL,
    guard boolean NOT NULL
);
--
--

ALTER TABLE ficheequipiers
    ADD UNIQUE (IdPersonne);
ALTER TABLE ficheequipiers
    ADD FOREIGN KEY (IdPersonne, guard) REFERENCES personnes (Id, IsTemp);
ALTER TABLE ficheequipiers
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;
ALTER TABLE ficheequipiers
    ADD CONSTRAINT Recommandation_gomacro CHECK (gomacro_validate_json_pers_Recommandation (Recommandation));
ALTER TABLE ficheequipiers
    ALTER COLUMN guard SET DEFAULT FALSE;
ALTER TABLE ficheequipiers
    ADD CHECK (guard = FALSE);
--
--

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
    '',
    '{"Nom": "", "Prenom": "", "Mail":"", "Tel": ""}',
    '',
    '',
    '',
    FALSE,
    FALSE
FROM
    personnes;
-- Cleanup
ALTER TABLE personnes
    DROP COLUMN DepartementNaissance;
ALTER TABLE personnes
    DROP COLUMN VilleNaissance;
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
--
--

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_Recommandation (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean;
BEGIN
    IF jsonb_typeof(data) != 'object' THEN
        RETURN FALSE;
    END IF;
    is_valid := (
        SELECT
            bool_and(KEY IN ('Nom', 'Prenom', 'Mail', 'Tel'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Nom')
        AND gomacro_validate_json_string (data -> 'Prenom')
        AND gomacro_validate_json_string (data -> 'Mail')
        AND gomacro_validate_json_string (data -> 'Tel');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
--
--

COMMIT;

