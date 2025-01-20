-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE personnes (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    NomJeuneFille text NOT NULL,
    Prenom text NOT NULL,
    DateNaissance timestamp(0) with time zone NOT NULL,
    VilleNaissance text NOT NULL,
    DepartementNaissance text NOT NULL,
    Sexe integer CHECK (Sexe IN (2, 1)) NOT NULL,
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
    FicheSanitaire jsonb NOT NULL,
    IsTemp boolean NOT NULL
);

-- constraints
CREATE OR REPLACE FUNCTION gomacro_validate_json_array_string (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) = 'null' THEN
        RETURN TRUE;
    END IF;
    IF jsonb_typeof(data) != 'array' THEN
        RETURN FALSE;
    END IF;
    IF jsonb_array_length(data) = 0 THEN
        RETURN TRUE;
    END IF;
    RETURN (
        SELECT
            bool_and(gomacro_validate_json_string (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

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

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_FicheSanitaire (data jsonb)
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
            bool_and(key IN ('traitement_medical', 'maladies', 'allergies', 'difficultes_sante', 'recommandations', 'handicap', 'tel', 'medecin', 'last_modif', 'mails'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'traitement_medical')
        AND gomacro_validate_json_pers_Maladies (data -> 'maladies')
        AND gomacro_validate_json_pers_Allergies (data -> 'allergies')
        AND gomacro_validate_json_string (data -> 'difficultes_sante')
        AND gomacro_validate_json_string (data -> 'recommandations')
        AND gomacro_validate_json_boolean (data -> 'handicap')
        AND gomacro_validate_json_string (data -> 'tel')
        AND gomacro_validate_json_pers_Medecin (data -> 'medecin')
        AND gomacro_validate_json_string (data -> 'last_modif')
        AND gomacro_validate_json_array_string (data -> 'mails');
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

ALTER TABLE personnes
    ADD CONSTRAINT FicheSanitaire_gomacro CHECK (gomacro_validate_json_pers_FicheSanitaire (FicheSanitaire));

