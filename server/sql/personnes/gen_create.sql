-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE fichesanitaires (
    IdPersonne integer NOT NULL,
    TraitementMedical boolean NOT NULL,
    Maladies jsonb NOT NULL,
    Allergies jsonb NOT NULL,
    DifficultesSante text NOT NULL,
    Recommandations text NOT NULL,
    Handicap boolean NOT NULL,
    Tel text NOT NULL,
    Medecin jsonb NOT NULL,
    LastModif timestamp(0) with time zone NOT NULL,
    Mails text[],
    guard boolean NOT NULL
);

CREATE TABLE personnes (
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

-- constraints
ALTER TABLE personnes
    ADD UNIQUE (Id, IsTemp);

ALTER TABLE fichesanitaires
    ADD UNIQUE (IdPersonne);

ALTER TABLE fichesanitaires
    ADD FOREIGN KEY (IdPersonne, guard) REFERENCES personnes (Id, IsTemp);

ALTER TABLE fichesanitaires
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;

ALTER TABLE fichesanitaires
    ALTER COLUMN guard SET DEFAULT FALSE;

ALTER TABLE fichesanitaires
    ADD CHECK (guard = FALSE);

CREATE OR REPLACE FUNCTION gomacro_validate_json_boolean (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'boolean';
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a boolean', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_Allergies (data jsonb)
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
            bool_and(key IN ('Asthme', 'Alimentaires', 'Medicamenteuses', 'Autres', 'ConduiteATenir'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'Asthme')
        AND gomacro_validate_json_boolean (data -> 'Alimentaires')
        AND gomacro_validate_json_boolean (data -> 'Medicamenteuses')
        AND gomacro_validate_json_string (data -> 'Autres')
        AND gomacro_validate_json_string (data -> 'ConduiteATenir');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_Maladies (data jsonb)
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
            bool_and(key IN ('Rubeole', 'Varicelle', 'Angine', 'Oreillons', 'Scarlatine', 'Coqueluche', 'Otite', 'Rougeole', 'Rhumatisme'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'Rubeole')
        AND gomacro_validate_json_boolean (data -> 'Varicelle')
        AND gomacro_validate_json_boolean (data -> 'Angine')
        AND gomacro_validate_json_boolean (data -> 'Oreillons')
        AND gomacro_validate_json_boolean (data -> 'Scarlatine')
        AND gomacro_validate_json_boolean (data -> 'Coqueluche')
        AND gomacro_validate_json_boolean (data -> 'Otite')
        AND gomacro_validate_json_boolean (data -> 'Rougeole')
        AND gomacro_validate_json_boolean (data -> 'Rhumatisme');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_Medecin (data jsonb)
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
            bool_and(key IN ('Nom', 'Tel'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Nom')
        AND gomacro_validate_json_string (data -> 'Tel');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_Publicite (data jsonb)
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
            bool_and(key IN ('VersionPapier', 'PubHiver', 'PubEte', 'EchoRocher', 'Eonews'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'VersionPapier')
        AND gomacro_validate_json_boolean (data -> 'PubHiver')
        AND gomacro_validate_json_boolean (data -> 'PubEte')
        AND gomacro_validate_json_boolean (data -> 'EchoRocher')
        AND gomacro_validate_json_boolean (data -> 'Eonews');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_string (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'string';
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a string', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

ALTER TABLE fichesanitaires
    ADD CONSTRAINT Allergies_gomacro CHECK (gomacro_validate_json_pers_Allergies (Allergies));

ALTER TABLE fichesanitaires
    ADD CONSTRAINT Maladies_gomacro CHECK (gomacro_validate_json_pers_Maladies (Maladies));

ALTER TABLE fichesanitaires
    ADD CONSTRAINT Medecin_gomacro CHECK (gomacro_validate_json_pers_Medecin (Medecin));

ALTER TABLE personnes
    ADD CONSTRAINT Publicite_gomacro CHECK (gomacro_validate_json_pers_Publicite (Publicite));

