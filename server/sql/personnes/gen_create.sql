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
    Mails text[]
);

CREATE TABLE personnes (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    NomJeuneFille text NOT NULL,
    Prenom text NOT NULL,
    DateNaissance date NOT NULL,
    VilleNaissance text NOT NULL,
    DepartementNaissance text NOT NULL,
    Sexe integer CHECK (Sexe IN (0, 1, 2)) NOT NULL,
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
    Diplome integer CHECK (Diplome IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)) NOT NULL,
    Approfondissement integer CHECK (Approfondissement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    IsTemp boolean NOT NULL
);

-- constraints
ALTER TABLE fichesanitaires
    ADD UNIQUE (IdPersonne);

ALTER TABLE fichesanitaires
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;

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
            bool_and(key IN ('asthme', 'alimentaires', 'medicamenteuses', 'autres', 'conduite_a_tenir'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'asthme')
        AND gomacro_validate_json_boolean (data -> 'alimentaires')
        AND gomacro_validate_json_boolean (data -> 'medicamenteuses')
        AND gomacro_validate_json_string (data -> 'autres')
        AND gomacro_validate_json_string (data -> 'conduite_a_tenir');
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
            bool_and(key IN ('rubeole', 'varicelle', 'angine', 'oreillons', 'scarlatine', 'coqueluche', 'otite', 'rougeole', 'rhumatisme'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'rubeole')
        AND gomacro_validate_json_boolean (data -> 'varicelle')
        AND gomacro_validate_json_boolean (data -> 'angine')
        AND gomacro_validate_json_boolean (data -> 'oreillons')
        AND gomacro_validate_json_boolean (data -> 'scarlatine')
        AND gomacro_validate_json_boolean (data -> 'coqueluche')
        AND gomacro_validate_json_boolean (data -> 'otite')
        AND gomacro_validate_json_boolean (data -> 'rougeole')
        AND gomacro_validate_json_boolean (data -> 'rhumatisme');
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
            bool_and(key IN ('nom', 'tel'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'nom')
        AND gomacro_validate_json_string (data -> 'tel');
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

