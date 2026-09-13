-- v0.12.0
BEGIN;
--
--
CREATE OR REPLACE FUNCTION gomacro_validate_json_array_camp_Champ (data jsonb)
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
            bool_and(gomacro_validate_json_camp_Champ (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_array_camp_ChampReponse (data jsonb)
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
            bool_and(gomacro_validate_json_camp_ChampReponse (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_array_number (data jsonb)
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
            bool_and(gomacro_validate_json_number (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
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
CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_Champ (data jsonb)
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
            bool_and(KEY IN ('Titre', 'Description', 'Question'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Titre')
        AND gomacro_validate_json_string (data -> 'Description')
        AND gomacro_validate_json_camp_ChampQuestion (data -> 'Question');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_ChampQCM (data jsonb)
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
            bool_and(KEY IN ('Multiple', 'Propositions'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'Multiple')
        AND gomacro_validate_json_array_string (data -> 'Propositions');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_ChampQuestion (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) != 'object' OR jsonb_typeof(data -> 'Kind') != 'string' OR jsonb_typeof(data -> 'Data') = 'null' THEN
        RETURN FALSE;
    END IF;
    CASE WHEN data ->> 'Kind' = 'ChampQCM' THEN
        RETURN gomacro_validate_json_camp_ChampQCM (data -> 'Data');
    WHEN data ->> 'Kind' = 'ChampTexte' THEN
        RETURN gomacro_validate_json_camp_ChampTexte (data -> 'Data');
    ELSE
        RETURN FALSE;
    END CASE;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_ChampReponse (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) != 'object' OR jsonb_typeof(data -> 'Kind') != 'string' OR jsonb_typeof(data -> 'Data') = 'null' THEN
        RETURN FALSE;
    END IF;
    CASE WHEN data ->> 'Kind' = 'ChampReponseQCM' THEN
        RETURN gomacro_validate_json_array_number (data -> 'Data');
    WHEN data ->> 'Kind' = 'ChampReponseTexte' THEN
        RETURN gomacro_validate_json_string (data -> 'Data');
    ELSE
        RETURN FALSE;
    END CASE;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_ChampTexte (data jsonb)
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
            bool_and(KEY IN ('MultiLignes'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'MultiLignes');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
CREATE OR REPLACE FUNCTION gomacro_validate_json_number (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number';
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a number', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;
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
--
--
CREATE TABLE forms (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    Nom text NOT NULL,
    Introduction text NOT NULL,
    Champs jsonb NOT NULL
);
CREATE TABLE participant_forms (
    IdParticipant integer NOT NULL,
    IdForm integer NOT NULL,
    IdCamp integer NOT NULL,
    Reponses jsonb NOT NULL
);
ALTER TABLE forms
    ADD UNIQUE (Id, IdCamp);
ALTER TABLE forms
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;
ALTER TABLE participant_forms
    ADD UNIQUE (IdParticipant, IdForm);
ALTER TABLE participant_forms
    ADD FOREIGN KEY (IdParticipant, IdCamp) REFERENCES participants (Id, IdCamp) ON DELETE CASCADE;
ALTER TABLE participant_forms
    ADD FOREIGN KEY (IdForm, IdCamp) REFERENCES forms (Id, IdCamp) ON DELETE CASCADE;
ALTER TABLE participant_forms
    ADD FOREIGN KEY (IdParticipant) REFERENCES participants;
ALTER TABLE participant_forms
    ADD FOREIGN KEY (IdForm) REFERENCES forms;
ALTER TABLE participant_forms
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;
ALTER TABLE participant_forms
    ADD CONSTRAINT Reponses_gomacro CHECK (gomacro_validate_json_array_camp_ChampReponse (Reponses));
--
COMMIT;

