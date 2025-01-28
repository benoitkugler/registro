-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE dossiers (
    Id serial PRIMARY KEY,
    IdPersonne integer NOT NULL,
    CopiesMails text[],
    LastConnection timestamp(0) with time zone NOT NULL,
    IsValidated boolean NOT NULL,
    PartageAdressesOK boolean NOT NULL
);

CREATE TABLE paiements (
    Id serial PRIMARY KEY,
    IdDossier integer NOT NULL,
    IsAcompte boolean NOT NULL,
    IsRemboursement boolean NOT NULL,
    Montant Montant NOT NULL,
    Payeur text NOT NULL,
    Mode smallint CHECK (Mode IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Date date NOT NULL,
    Label text NOT NULL,
    Details text NOT NULL
);

CREATE TABLE participants (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IdPersonne integer NOT NULL,
    IdDossier integer NOT NULL,
    ListeAttente smallint CHECK (ListeAttente IN (0, 1, 2, 3, 4)) NOT NULL,
    Remises jsonb NOT NULL,
    QuotientFamilial integer NOT NULL,
    Details text NOT NULL,
    Bus smallint CHECK (Bus IN (0, 1, 2, 3)) NOT NULL
);

CREATE TABLE sondages (
    IdSondage integer NOT NULL,
    IdCamp integer NOT NULL,
    IdDossier integer NOT NULL,
    Modified timestamp(0) with time zone NOT NULL,
    InfosAvantSejour smallint CHECK (InfosAvantSejour IN (0, 1, 2, 3, 4)) NOT NULL,
    InfosPendantSejour smallint CHECK (InfosPendantSejour IN (0, 1, 2, 3, 4)) NOT NULL,
    Hebergement smallint CHECK (Hebergement IN (0, 1, 2, 3, 4)) NOT NULL,
    Activites smallint CHECK (Activites IN (0, 1, 2, 3, 4)) NOT NULL,
    Theme smallint CHECK (Theme IN (0, 1, 2, 3, 4)) NOT NULL,
    Nourriture smallint CHECK (Nourriture IN (0, 1, 2, 3, 4)) NOT NULL,
    Hygiene smallint CHECK (Hygiene IN (0, 1, 2, 3, 4)) NOT NULL,
    Ambiance smallint CHECK (Ambiance IN (0, 1, 2, 3, 4)) NOT NULL,
    Ressenti smallint CHECK (Ressenti IN (0, 1, 2, 3, 4)) NOT NULL,
    MessageEnfant text NOT NULL,
    MessageResponsable text NOT NULL
);

-- constraints
ALTER TABLE participants
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE participants
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;

ALTER TABLE participants
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE dossiers
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes;

ALTER TABLE paiements
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE sondages
    ADD UNIQUE (IdCamp, IdDossier);

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_Currency (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a camp_Currency', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_Montant (data jsonb)
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
            bool_and(key IN ('Cent', 'Currency'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'Cent')
        AND gomacro_validate_json_camp_Currency (data -> 'Currency');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_doss_Remises (data jsonb)
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
            bool_and(key IN ('ReducEquipiers', 'ReducEnfants', 'ReducSpeciale'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'ReducEquipiers')
        AND gomacro_validate_json_number (data -> 'ReducEnfants')
        AND gomacro_validate_json_camp_Montant (data -> 'ReducSpeciale');
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

ALTER TABLE participants
    ADD CONSTRAINT Remises_gomacro CHECK (gomacro_validate_json_doss_Remises (Remises));

