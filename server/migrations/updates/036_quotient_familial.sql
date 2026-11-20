-- v0.13.1
-- Grilles de quotient familial personnalisées : la grille devient une liste
-- de tranches, stockée au format JSONB. Une liste vide désactive l'option.

BEGIN;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_TrancheQuotient (data jsonb)
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
            bool_and(KEY IN ('QuotientMin', 'Pourcentage'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'QuotientMin')
        AND gomacro_validate_json_number (data -> 'Pourcentage');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_camp_TrancheQuotient (data jsonb)
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
            bool_and(gomacro_validate_json_camp_TrancheQuotient (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

ALTER TABLE camps DROP CONSTRAINT IF EXISTS camps_optionquotientfamilial_check;

ALTER TABLE camps
    ALTER COLUMN OptionQuotientFamilial TYPE jsonb
    USING CASE
        WHEN OptionQuotientFamilial[4] = 100 THEN jsonb_build_array(
            jsonb_build_object('QuotientMin', 1, 'Pourcentage', OptionQuotientFamilial[1]),
            jsonb_build_object('QuotientMin', 360, 'Pourcentage', OptionQuotientFamilial[2]),
            jsonb_build_object('QuotientMin', 565, 'Pourcentage', OptionQuotientFamilial[3]),
            jsonb_build_object('QuotientMin', 715, 'Pourcentage', OptionQuotientFamilial[4])
        )
        ELSE '[]'::jsonb
    END;

ALTER TABLE camps
    ADD CONSTRAINT OptionQuotientFamilial_gomacro CHECK (gomacro_validate_json_array_camp_TrancheQuotient (OptionQuotientFamilial));

ROLLBACK;
