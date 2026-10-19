-- v0.12.3
-- Simplifie les options sur le prix
BEGIN;
  -- 
 
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
            bool_and(KEY IN ('Active', 'InscriptionRapide', 'PrixParJour'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_camp_OptionPrixKind (data -> 'Active')
        AND gomacro_validate_json_camp_InscriptionRapide (data -> 'InscriptionRapide')
        AND gomacro_validate_json_number (data -> 'PrixParJour');
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
    AND data::int IN (0, 1, 2);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a camp_OptionPrixKind', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_OptionPrixParticipant (data jsonb)
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
            bool_and(KEY IN ('Jour'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_array_number (data -> 'Jour');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_Remises (data jsonb)
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
            bool_and(KEY IN ('Equipiers', 'Famille', 'Speciale'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'Equipiers')
        AND gomacro_validate_json_number (data -> 'Famille')
        AND gomacro_validate_json_doss_Montant (data -> 'Speciale');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

UPDATE camps SET OptionPrix = ((OptionPrix - 'Statuts') - 'Jours') || '{ "PrixParJour" : 0 }';
UPDATE participants SET OptionPrix = '{"Jour":[]}';

COMMIT;
