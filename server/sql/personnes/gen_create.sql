-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
DROP TYPE IF EXISTS Nationnalite;

CREATE TYPE Nationnalite AS (
    IsSuisse boolean
);

DROP TYPE IF EXISTS Publicite;

CREATE TYPE Publicite AS (
    VersionPapier boolean,
    PubHiver boolean,
    PubEte boolean,
    EchoRocher boolean,
    Eonews boolean
);

CREATE TABLE fichesanitaires (
    IdPersonne integer NOT NULL,
    DifficultesSante text NOT NULL,
    AllergiesAlimentaires text NOT NULL,
    TraitementMedical text NOT NULL,
    Medecin jsonb NOT NULL,
    AutreContact jsonb NOT NULL,
    Modified timestamp(0) with time zone NOT NULL,
    Owners text[],
    guard boolean NOT NULL
);

CREATE TABLE personnes (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    Prenom text NOT NULL,
    Sexe smallint CHECK (Sexe IN (0, 1, 2)) NOT NULL,
    DateNaissance date NOT NULL,
    VilleNaissance text NOT NULL,
    DepartementNaissance text NOT NULL,
    Nationnalite Nationnalite NOT NULL,
    Tels text[],
    Mail text NOT NULL,
    Adresse text NOT NULL,
    CodePostal text NOT NULL,
    Ville text NOT NULL,
    Pays text NOT NULL,
    NomJeuneFille text NOT NULL,
    Profession text NOT NULL,
    Etudiant boolean NOT NULL,
    Fonctionnaire boolean NOT NULL,
    Diplome smallint CHECK (Diplome IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)) NOT NULL,
    Approfondissement smallint CHECK (Approfondissement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Publicite Publicite NOT NULL,
    CharteAccepted timestamp(0) with time zone NOT NULL,
    IsTemp boolean NOT NULL
);

-- constraints
ALTER TABLE personnes
    ADD UNIQUE (Id, IsTemp);

ALTER TABLE fichesanitaires
    ADD UNIQUE (IdPersonne);

ALTER TABLE fichesanitaires
    ADD FOREIGN KEY (IdPersonne, guard) REFERENCES personnes (Id, IsTemp);

ALTER TABLE fichesanitaires
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;

ALTER TABLE fichesanitaires
    ALTER COLUMN guard SET DEFAULT FALSE;

ALTER TABLE fichesanitaires
    ADD CHECK (guard = FALSE);

CREATE OR REPLACE FUNCTION gomacro_validate_json_pers_NomTel (data jsonb)
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
            bool_and(key IN ('Nom', 'Tel'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'Nom')
        AND gomacro_validate_json_string (data -> 'Tel');
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

ALTER TABLE fichesanitaires
    ADD CONSTRAINT AutreContact_gomacro CHECK (gomacro_validate_json_pers_NomTel (AutreContact));

ALTER TABLE fichesanitaires
    ADD CONSTRAINT Medecin_gomacro CHECK (gomacro_validate_json_pers_NomTel (Medecin));

