-- v0.13.0 : Compta App !

BEGIN;

CREATE TABLE compta_budgets (
    IdCompta integer NOT NULL,
    SimulatedInscrits integer NOT NULL,
    SimulatedEquipiers integer NOT NULL,
    StartMealInscrits smallint CHECK (StartMealInscrits IN (0, 1, 2, 3, 4)) NOT NULL,
    EndMealInscrits smallint CHECK (EndMealInscrits IN (0, 1, 2, 3, 4)) NOT NULL,
    StartMealEquipiers smallint CHECK (StartMealEquipiers IN (0, 1, 2, 3, 4)) NOT NULL,
    EndMealEquipiers smallint CHECK (EndMealEquipiers IN (0, 1, 2, 3, 4)) NOT NULL,
    OffsetStartMealEquipiers integer NOT NULL,
    OffsetEndMealEquipiers integer NOT NULL,
    ExternalHost jsonb NOT NULL,
    Activities jsonb NOT NULL,
    MaterielPeda Montant NOT NULL,
    ProduitPharma Montant NOT NULL,
    TeamTrips Montant NOT NULL,
    ExternalCarRent Montant NOT NULL,
    AssoCarKms integer[] CHECK (array_length(AssoCarKms, 1) = 3) NOT NULL
);

CREATE TABLE compta_camps (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    FileStatus smallint CHECK (FileStatus IN (0, 1, 2)) NOT NULL,
    State smallint CHECK (State IN (0, 1, 2, 3)) NOT NULL,
    LastSentMailAdmin timestamp(0) with time zone NOT NULL,
    LastSentMailDirecteur timestamp(0) with time zone NOT NULL,
    Account smallint CHECK (Account IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)) NOT NULL,
    Fraismdr jsonb NOT NULL
);

CREATE TABLE compta_configs (
    Id serial PRIMARY KEY,
    Year integer NOT NULL,
    MealsCost integer[] CHECK (array_length(MealsCost, 1) = 5) NOT NULL,
    SmallMealThreshold integer NOT NULL,
    SmallMealMajoration integer NOT NULL,
    AssoCars integer[] CHECK (array_length(AssoCars, 1) = 3) NOT NULL,
    ChauffageMDR integer NOT NULL,
    EntretienMDR integer NOT NULL,
    MaterielMDR integer NOT NULL
);

CREATE TABLE compta_files (
    Id serial PRIMARY KEY,
    Name text NOT NULL,
    DriveID text NOT NULL,
    DriveIDMiniature text NOT NULL
);

CREATE TABLE compta_records (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IsReceipe boolean NOT NULL,
    SpentBy integer,
    Reviews jsonb NOT NULL,
    IsValidated boolean NOT NULL,
    Account smallint CHECK (Account IN (0, 1)) NOT NULL
);

CREATE TABLE compta_record_buss (
    IdCamp integer NOT NULL,
    IdSubrecord integer NOT NULL
);

CREATE TABLE compta_subrecords (
    Id serial PRIMARY KEY,
    IdRecord integer NOT NULL,
    Date date NOT NULL,
    Description text NOT NULL,
    Mode smallint CHECK (Mode IN (0, 1, 2, 3, 4)) NOT NULL,
    Serial text NOT NULL,
    Currency smallint CHECK (Currency IN (0, 1)) NOT NULL,
    Montants jsonb NOT NULL,
    IdFile integer
);

-- constraints
ALTER TABLE compta_camps
    ADD UNIQUE (IdCamp);

ALTER TABLE compta_camps
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE compta_records
    ADD UNIQUE (Id, IdCamp);

ALTER TABLE compta_records
    ADD FOREIGN KEY (IdCamp) REFERENCES compta_camps ON DELETE CASCADE;

ALTER TABLE compta_records
    ADD FOREIGN KEY (SpentBy) REFERENCES equipiers;

ALTER TABLE compta_subrecords
    ADD FOREIGN KEY (IdRecord) REFERENCES compta_records ON DELETE CASCADE;

ALTER TABLE compta_subrecords
    ADD FOREIGN KEY (IdFile) REFERENCES compta_files ON DELETE SET NULL;

ALTER TABLE compta_record_buss
    ADD UNIQUE (IdSubrecord);

ALTER TABLE compta_record_buss
    ADD FOREIGN KEY (IdCamp) REFERENCES compta_camps ON DELETE CASCADE;

ALTER TABLE compta_record_buss
    ADD FOREIGN KEY (IdSubrecord) REFERENCES compta_subrecords ON DELETE CASCADE;

ALTER TABLE compta_budgets
    ADD UNIQUE (IdCompta);

ALTER TABLE compta_budgets
    ADD FOREIGN KEY (IdCompta) REFERENCES compta_camps ON DELETE CASCADE;

ALTER TABLE compta_configs
    ADD UNIQUE (Year);

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

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_comp_NamedCost (data jsonb)
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
            bool_and(gomacro_validate_json_comp_NamedCost (value))
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_FraisMDRACVE (data jsonb)
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_comp_NamedCost (data jsonb)
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
            bool_and(KEY IN ('Name', 'Cost'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Name')
        AND gomacro_validate_json_doss_Montant (data -> 'Cost');
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_doss_Currency (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a doss_Currency', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_doss_Montant (data jsonb)
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
            bool_and(KEY IN ('Cent', 'Currency'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'Cent')
        AND gomacro_validate_json_doss_Currency (data -> 'Currency');
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

ALTER TABLE compta_subrecords
    ADD CONSTRAINT Montants_gomacro CHECK (gomacro_validate_json_array_comp_MontantPart (Montants));

ALTER TABLE compta_budgets
    ADD CONSTRAINT Activities_gomacro CHECK (gomacro_validate_json_array_comp_NamedCost (Activities));

ALTER TABLE compta_budgets
    ADD CONSTRAINT ExternalHost_gomacro CHECK (gomacro_validate_json_array_comp_NamedCost (ExternalHost));

ALTER TABLE compta_records
    ADD CONSTRAINT Reviews_gomacro CHECK (gomacro_validate_json_array_comp_Review (Reviews));

ALTER TABLE compta_camps
    ADD CONSTRAINT Fraismdr_gomacro CHECK (gomacro_validate_json_comp_FraisMDRACVE (Fraismdr));


-- default compta config

INSERT INTO compta_configs (Year, -- integer
    MealsCost, -- integer[]
    SmallMealThreshold, -- integer
    SmallMealMajoration, -- integer
    AssoCars, -- integer[]
    ChauffageMDR, -- integer
    EntretienMDR, -- integer
    MaterielMDR -- integer
 )
    VALUES (2020, '{50, 200, 50, 200, 70}', 25, 10, '{150, 80, 50}', 15000, 20, 100);

COMMIT;