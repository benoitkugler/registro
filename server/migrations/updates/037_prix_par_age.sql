-- v0.13.1
-- Nouvelle option sur le prix : un barème par âge du participant.
-- Le barème est stocké dans le JSONB OptionPrix des camps, il n'y a donc
-- pas de nouvelle colonne ni de migration de données. Les anciennes valeurs
-- (sans la clé PrixParAge) restent valides.

BEGIN;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_PrixAgeTranche (data jsonb)
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
            bool_and(KEY IN ('AgeMax', 'Prix'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'AgeMax')
        AND gomacro_validate_json_number (data -> 'Prix');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_camp_PrixAgeTranche (data jsonb)
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
            bool_and(gomacro_validate_json_camp_PrixAgeTranche (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_OptionPrixCamp (data jsonb)
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
            bool_and(KEY IN ('Active', 'InscriptionRapide', 'PrixParJour', 'PrixParAge'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_camp_OptionPrixKind (data -> 'Active')
        AND gomacro_validate_json_camp_InscriptionRapide (data -> 'InscriptionRapide')
        AND gomacro_validate_json_number (data -> 'PrixParJour')
        AND gomacro_validate_json_array_camp_PrixAgeTranche (data -> 'PrixParAge');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_OptionPrixKind (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2, 3);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a camp_OptionPrixKind', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

COMMIT;
