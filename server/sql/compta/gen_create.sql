-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE compta_bankentrys (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IsReceipe boolean NOT NULL,
    SpentBy integer,
    Reviews jsonb NOT NULL,
    IsValidated boolean NOT NULL,
    Account smallint CHECK (Account IN (0, 1)) NOT NULL
);

CREATE TABLE compta_camps (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    FileStatus smallint CHECK (FileStatus IN (0, 1, 2)) NOT NULL,
    State smallint CHECK (State IN (0, 1, 2, 3)) NOT NULL,
    LastSentMailAdmin timestamp(0) with time zone NOT NULL,
    LastSentMailDirecteur timestamp(0) with time zone NOT NULL
);

CREATE TABLE compta_camp_acves (
    IdComptaCamp integer NOT NULL,
    Account smallint CHECK (Account IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)) NOT NULL,
    Fraismdr jsonb NOT NULL
);

CREATE TABLE compta_files (
    Id serial PRIMARY KEY,
    Name text NOT NULL,
    DriveID text NOT NULL,
    DriveIDMiniature text NOT NULL
);

CREATE TABLE compta_records (
    Id serial PRIMARY KEY,
    IdBankentry integer NOT NULL,
    Date date NOT NULL,
    Description text NOT NULL,
    Mode smallint CHECK (Mode IN (0, 1, 2, 3, 4)) NOT NULL,
    Serial text NOT NULL,
    Currency smallint CHECK (Currency IN (1, 2)) NOT NULL,
    Montants jsonb NOT NULL,
    IdFile integer
);

CREATE TABLE compta_record_buss (
    IdCamp integer NOT NULL,
    IdRecord integer NOT NULL
);

-- constraints
ALTER TABLE compta_camps
    ADD UNIQUE (IdCamp);

ALTER TABLE compta_camps
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE compta_camp_acves
    ADD FOREIGN KEY (IdComptaCamp) REFERENCES compta_camps ON DELETE CASCADE;

ALTER TABLE compta_bankentrys
    ADD UNIQUE (Id, IdCamp);

ALTER TABLE compta_bankentrys
    ADD FOREIGN KEY (IdCamp) REFERENCES compta_camps ON DELETE CASCADE;

ALTER TABLE compta_bankentrys
    ADD FOREIGN KEY (SpentBy) REFERENCES equipiers;

ALTER TABLE compta_records
    ADD FOREIGN KEY (IdBankentry) REFERENCES compta_bankentrys ON DELETE CASCADE;

ALTER TABLE compta_records
    ADD FOREIGN KEY (IdFile) REFERENCES compta_files ON DELETE SET NULL;

ALTER TABLE compta_record_buss
    ADD UNIQUE (IdRecord);

ALTER TABLE compta_record_buss
    ADD FOREIGN KEY (IdCamp) REFERENCES compta_camps ON DELETE CASCADE;

ALTER TABLE compta_record_buss
    ADD FOREIGN KEY (IdRecord) REFERENCES compta_records ON DELETE CASCADE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_comp_Bus (data jsonb)
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
            bool_and(gomacro_validate_json_comp_Bus (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_comp_MontantPart (data jsonb)
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
            bool_and(gomacro_validate_json_comp_MontantPart (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_comp_Review (data jsonb)
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
            bool_and(gomacro_validate_json_comp_Review (value))
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_Bus (data jsonb)
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
            bool_and(KEY IN ('Name', 'Kind', 'KmStart', 'KmEnd'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Name')
        AND gomacro_validate_json_comp_CarKind (data -> 'Kind')
        AND gomacro_validate_json_number (data -> 'KmStart')
        AND gomacro_validate_json_number (data -> 'KmEnd');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_CarKind (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2, 3);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a comp_CarKind', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_FraismdrACVE (data jsonb)
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
            bool_and(KEY IN ('Bus'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_array_comp_Bus (data -> 'Bus');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_MontantPart (data jsonb)
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
            bool_and(KEY IN ('Cent', 'Kind'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'Cent')
        AND gomacro_validate_json_comp_Ventilation (data -> 'Kind');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_Review (data jsonb)
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
            bool_and(KEY IN ('IsAdmin', 'Time', 'Message'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'IsAdmin')
        AND gomacro_validate_json_string (data -> 'Time')
        AND gomacro_validate_json_string (data -> 'Message');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_Ventilation (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a comp_Ventilation', data;
    END IF;
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

ALTER TABLE compta_records
    ADD CONSTRAINT Montants_gomacro CHECK (gomacro_validate_json_array_comp_MontantPart (Montants));

ALTER TABLE compta_bankentrys
    ADD CONSTRAINT Reviews_gomacro CHECK (gomacro_validate_json_array_comp_Review (Reviews));

ALTER TABLE compta_camp_acves
    ADD CONSTRAINT Fraismdr_gomacro CHECK (gomacro_validate_json_comp_FraismdrACVE (Fraismdr));

