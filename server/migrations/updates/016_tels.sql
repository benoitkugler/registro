-- v0.12 : limiter les numéros de tel à 2
-- check no more 2 numbers are actually used

SELECT
    id,
    nom,
    prenom,
    tels,
    array_length(tels, 1)
FROM
    personnes
WHERE
    array_length(tels, 1) > 2;

SELECT
    id,
    responsable ->> 'Nom',
    responsable ->> 'Prenom',
    responsable -> 'Tels',
    jsonb_array_length(responsable -> 'Tels')
FROM
    inscriptions
WHERE
    jsonb_array_length(responsable -> 'Tels') > 2;

BEGIN;
CREATE OR REPLACE FUNCTION __migration_tels (tels text[])
    RETURNS text[]
    AS $$
BEGIN
    RETURN (
        CASE WHEN tels IS NULL
            OR array_length(tels, 1) = 0
            OR array_length(tels, 1) IS NULL THEN
            '{"", ""}' WHEN array_length(tels, 1) = 1 THEN
            tels || '{""}'
        ELSE
            tels[array_length(tels, 1) - 1:] -- 1-based
        END);
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
--
--

CREATE OR REPLACE FUNCTION __migration_tels_jsonb (tels jsonb)
    RETURNS jsonb
    AS $$
BEGIN
    RETURN to_jsonb (__migration_tels (ARRAY (
                SELECT
                    jsonb_array_elements_text(tels))));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
--
--

UPDATE
    personnes
SET
    Tels = __migration_tels (Tels);
ALTER TABLE personnes
    ALTER COLUMN Tels SET NOT NULL;
ALTER TABLE personnes
    ADD CONSTRAINT personnes_tels_check CHECK (array_length(Tels, 1) = 2);
--
--

UPDATE
    inscriptions
SET
    responsable = jsonb_set(responsable, '{Tels}', __migration_tels_jsonb (responsable -> 'Tels'));
--
--

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_2_string (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) != 'array' THEN
        RETURN FALSE;
    END IF;
    RETURN (
        SELECT
            bool_and(gomacro_validate_json_string (value))
        FROM
            jsonb_array_elements(data))
        AND jsonb_array_length(data) = 2;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_insc_ResponsableLegal (data jsonb)
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
            bool_and(KEY IN ('Nom', 'Prenom', 'DateNaissance', 'Sexe', 'Mail', 'Tels', 'Adresse', 'CodePostal', 'Ville', 'Pays'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Nom')
        AND gomacro_validate_json_string (data -> 'Prenom')
        AND gomacro_validate_json_string (data -> 'DateNaissance')
        AND gomacro_validate_json_pers_Sexe (data -> 'Sexe')
        AND gomacro_validate_json_string (data -> 'Mail')
        AND gomacro_validate_json_array_2_string (data -> 'Tels')
        AND gomacro_validate_json_string (data -> 'Adresse')
        AND gomacro_validate_json_string (data -> 'CodePostal')
        AND gomacro_validate_json_string (data -> 'Ville')
        AND gomacro_validate_json_string (data -> 'Pays');
    IF is_valid = FALSE THEN
        RAISE 'not valid %', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
-- refresh the contraint for existing tables
ALTER TABLE inscriptions
    DROP CONSTRAINT Responsable_gomacro;
ALTER TABLE inscriptions
    ADD CONSTRAINT Responsable_gomacro CHECK (gomacro_validate_json_insc_ResponsableLegal (Responsable));
--
--

DROP FUNCTION __migration_tels;
DROP FUNCTION __migration_tels_jsonb;
COMMIT;

