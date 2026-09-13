-- v0.12
-- ajoute une option remise pour inscription rapide
BEGIN;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_InscriptionRapide (data jsonb)
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
            bool_and(KEY IN ('Limite', 'Prix'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Limite')
        AND gomacro_validate_json_number (data -> 'Prix');
    RETURN is_valid;
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
            bool_and(KEY IN ('Active', 'InscriptionRapide', 'Statuts', 'Jours'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_camp_OptionPrixKind (data -> 'Active')
        AND gomacro_validate_json_camp_InscriptionRapide (data -> 'InscriptionRapide')
        AND gomacro_validate_json_array_camp_PrixParStatut (data -> 'Statuts')
        AND gomacro_validate_json_array_number (data -> 'Jours');
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

UPDATE camps SET OptionPrix = OptionPrix || '{"InscriptionRapide" : {}}';

COMMIT;

