-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE inscriptions (
    Id serial PRIMARY KEY,
    IdTaux integer NOT NULL,
    Responsable jsonb NOT NULL,
    Message text NOT NULL,
    CopiesMails text[],
    PartageAdressesOK boolean NOT NULL,
    DemandeFondSoutien boolean NOT NULL,
    DateHeure timestamp(0) with time zone NOT NULL,
    ConfirmedAsDossier integer
);

CREATE TABLE inscription_participants (
    IdInscription integer NOT NULL,
    IdCamp integer NOT NULL,
    IdTaux integer NOT NULL,
    Nom text NOT NULL,
    Prenom text NOT NULL,
    DateNaissance date NOT NULL,
    Sexe smallint CHECK (Sexe IN (0, 1, 2)) NOT NULL,
    Nationnalite Nationnalite NOT NULL
);

-- constraints
ALTER TABLE inscriptions
    ADD UNIQUE (Id, IdTaux);

ALTER TABLE inscriptions
    ADD FOREIGN KEY (IdTaux) REFERENCES tauxs;

ALTER TABLE inscriptions
    ADD FOREIGN KEY (ConfirmedAsDossier) REFERENCES dossiers ON DELETE SET NULL;

ALTER TABLE inscription_participants
    ADD FOREIGN KEY (IdCamp, IdTaux) REFERENCES camps (Id, IdTaux) ON DELETE CASCADE;

ALTER TABLE inscription_participants
    ADD FOREIGN KEY (IdInscription, IdTaux) REFERENCES inscriptions (Id, IdTaux) ON DELETE CASCADE;

ALTER TABLE inscription_participants
    ADD FOREIGN KEY (IdInscription) REFERENCES inscriptions ON DELETE CASCADE;

ALTER TABLE inscription_participants
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE inscription_participants
    ADD FOREIGN KEY (IdTaux) REFERENCES tauxs;

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
            bool_and(key IN ('Nom', 'Prenom', 'DateNaissance', 'Sexe', 'Mail', 'Tels', 'Adresse', 'CodePostal', 'Ville', 'Pays'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Nom')
        AND gomacro_validate_json_string (data -> 'Prenom')
        AND gomacro_validate_json_string (data -> 'DateNaissance')
        AND gomacro_validate_json_pers_Sexe (data -> 'Sexe')
        AND gomacro_validate_json_string (data -> 'Mail')
        AND gomacro_validate_json_array_string (data -> 'Tels')
        AND gomacro_validate_json_string (data -> 'Adresse')
        AND gomacro_validate_json_string (data -> 'CodePostal')
        AND gomacro_validate_json_string (data -> 'Ville')
        AND gomacro_validate_json_string (data -> 'Pays');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_Sexe (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a pers_Sexe', data;
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

ALTER TABLE inscriptions
    ADD CONSTRAINT Responsable_gomacro CHECK (gomacro_validate_json_insc_ResponsableLegal (Responsable));

