-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TYPE Montant AS (
    Cent integer,
    Currency smallint
);

CREATE TABLE camps (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    DateDebut date NOT NULL,
    Duree integer NOT NULL,
    Agrement text NOT NULL,
    Prix Montant NOT NULL
);

CREATE TABLE equipiers (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IdPersonne integer NOT NULL,
    Roles jsonb NOT NULL,
    Presence jsonb NOT NULL,
    Invitation smallint CHECK (Invitation IN (0, 1, 2)) NOT NULL,
    AccepteCharte boolean
);

CREATE TABLE imagelettres (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    Filename text NOT NULL,
    Content bytea NOT NULL
);

CREATE TABLE lettredirecteurs (
    IdCamp integer NOT NULL,
    Html text NOT NULL,
    UseCoordCentre boolean NOT NULL,
    ShowAdressePostale boolean NOT NULL,
    ColorCoord text NOT NULL
);

-- constraints
ALTER TABLE lettredirecteurs
    ADD UNIQUE (IdCamp);

ALTER TABLE equipiers
    ADD UNIQUE (IdCamp, IdPersonne);

ALTER TABLE equipiers
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE equipiers
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_camp_Role (data jsonb)
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
            bool_and(gomacro_validate_json_camp_Role (value))
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_Role (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a camp_Role', data;
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_shar_OptionnalPlage (data jsonb)
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
            bool_and(key IN ('From', 'Duree', 'Active'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'From')
        AND gomacro_validate_json_number (data -> 'Duree')
        AND gomacro_validate_json_boolean (data -> 'Active');
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

ALTER TABLE equipiers
    ADD CONSTRAINT Roles_gomacro CHECK (gomacro_validate_json_array_camp_Role (Roles));

ALTER TABLE equipiers
    ADD CONSTRAINT Presence_gomacro CHECK (gomacro_validate_json_shar_OptionnalPlage (Presence));

