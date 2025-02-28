-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE aides (
    Id serial PRIMARY KEY,
    IdStructureaide integer NOT NULL,
    IdParticipant integer NOT NULL,
    Valide boolean NOT NULL,
    Valeur Montant NOT NULL,
    ParJour boolean NOT NULL,
    NbJoursMax integer NOT NULL
);

CREATE TABLE camps (
    Id serial PRIMARY KEY,
    IdTaux integer NOT NULL,
    Nom text NOT NULL,
    DateDebut date NOT NULL,
    Duree integer NOT NULL,
    Lieu text NOT NULL,
    Agrement text NOT NULL,
    Description text NOT NULL,
    Navette jsonb NOT NULL,
    Places integer NOT NULL,
    AgeMin integer NOT NULL,
    AgeMax integer NOT NULL,
    NeedEquilibreGF boolean NOT NULL,
    Ouvert boolean NOT NULL,
    Prix Montant NOT NULL,
    OptionPrix jsonb NOT NULL,
    OptionQuotientFamilial integer[] CHECK (array_length(OptionQuotientFamilial, 1) = 4) NOT NULL,
    Password text NOT NULL
);

CREATE TABLE equipiers (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IdPersonne integer NOT NULL,
    Roles smallint[],
    Presence jsonb NOT NULL,
    Invitation smallint CHECK (Invitation IN (0, 1, 2)) NOT NULL,
    AccepteCharte boolean
);

CREATE TABLE groupes (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    Nom text NOT NULL,
    Plage jsonb NOT NULL,
    Couleur text NOT NULL
);

CREATE TABLE groupe_participants (
    IdParticipant integer NOT NULL,
    IdGroupe integer NOT NULL,
    IdCamp integer NOT NULL,
    Manuel boolean NOT NULL
);

CREATE TABLE lettre_images (
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

CREATE TABLE participants (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IdPersonne integer NOT NULL,
    IdDossier integer NOT NULL,
    IdTaux integer NOT NULL,
    Statut smallint CHECK (Statut IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Remises jsonb NOT NULL,
    QuotientFamilial integer NOT NULL,
    OptionPrix jsonb NOT NULL,
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

CREATE TABLE structureaides (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    Immatriculation text NOT NULL,
    Adresse text NOT NULL,
    CodePostal text NOT NULL,
    Ville text NOT NULL,
    Telephone text NOT NULL,
    Info text NOT NULL
);

-- constraints
ALTER TABLE camps
    ADD UNIQUE (Id, IdTaux);

ALTER TABLE camps
    ADD FOREIGN KEY (IdTaux) REFERENCES tauxs;

ALTER TABLE lettredirecteurs
    ADD UNIQUE (IdCamp);

ALTER TABLE lettredirecteurs
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE participants
    ADD FOREIGN KEY (IdCamp, IdTaux) REFERENCES camps (Id, IdTaux);

ALTER TABLE participants
    ADD FOREIGN KEY (IdDossier, IdTaux) REFERENCES dossiers (Id, IdTaux) ON DELETE CASCADE;

ALTER TABLE participants
    ADD UNIQUE (Id, IdCamp);

ALTER TABLE participants
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE participants
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes;

ALTER TABLE participants
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE participants
    ADD FOREIGN KEY (IdTaux) REFERENCES tauxs;

ALTER TABLE groupes
    ADD UNIQUE (IdCamp, Nom);

ALTER TABLE groupes
    ADD UNIQUE (Id, IdCamp);

ALTER TABLE groupes
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE groupe_participants
    ADD UNIQUE (IdParticipant);

ALTER TABLE groupe_participants
    ADD UNIQUE (IdParticipant, IdCamp);

ALTER TABLE groupe_participants
    ADD FOREIGN KEY (IdParticipant, IdCamp) REFERENCES participants (Id, IdCamp) ON DELETE CASCADE;

ALTER TABLE groupe_participants
    ADD FOREIGN KEY (IdGroupe, IdCamp) REFERENCES groupes (Id, IdCamp) ON DELETE CASCADE;

ALTER TABLE groupe_participants
    ADD FOREIGN KEY (IdParticipant) REFERENCES participants ON DELETE CASCADE;

ALTER TABLE groupe_participants
    ADD FOREIGN KEY (IdGroupe) REFERENCES groupes ON DELETE CASCADE;

ALTER TABLE groupe_participants
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE equipiers
    ADD UNIQUE (IdCamp, IdPersonne);

CREATE UNIQUE INDEX ON equipiers (IdCamp)
WHERE
    1
    /* Role.Direction */
    = ANY (Roles);

ALTER TABLE equipiers
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE equipiers
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes ON DELETE CASCADE;

ALTER TABLE sondages
    ADD UNIQUE (IdCamp, IdDossier);

ALTER TABLE sondages
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE sondages
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE aides
    ADD FOREIGN KEY (IdStructureaide) REFERENCES structureaides;

ALTER TABLE aides
    ADD FOREIGN KEY (IdParticipant) REFERENCES participants ON DELETE CASCADE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_camp_PrixParStatut (data jsonb)
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
            bool_and(gomacro_validate_json_camp_PrixParStatut (value))
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_Navette (data jsonb)
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
            bool_and(key IN ('Actif', 'Commentaire'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_boolean (data -> 'Actif')
        AND gomacro_validate_json_string (data -> 'Commentaire');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

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
            bool_and(key IN ('Active', 'Statuts', 'Jours'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_camp_OptionPrixKind (data -> 'Active')
        AND gomacro_validate_json_array_camp_PrixParStatut (data -> 'Statuts')
        AND gomacro_validate_json_array_number (data -> 'Jours');
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
            bool_and(key IN ('IdStatut', 'Jour'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'IdStatut')
        AND gomacro_validate_json_array_number (data -> 'Jour');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_OptionnalPlage (data jsonb)
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_camp_PrixParStatut (data jsonb)
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
            bool_and(key IN ('Id', 'Prix', 'Label', 'Description'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'Id')
        AND gomacro_validate_json_number (data -> 'Prix')
        AND gomacro_validate_json_string (data -> 'Label')
        AND gomacro_validate_json_string (data -> 'Description');
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
            bool_and(key IN ('ReducEquipiers', 'ReducEnfants', 'ReducSpeciale'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'ReducEquipiers')
        AND gomacro_validate_json_number (data -> 'ReducEnfants')
        AND gomacro_validate_json_doss_Montant (data -> 'ReducSpeciale');
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
            bool_and(key IN ('Cent', 'Currency'))
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

CREATE OR REPLACE FUNCTION gomacro_validate_json_shar_Plage (data jsonb)
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
            bool_and(key IN ('From', 'Duree'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_string (data -> 'From')
        AND gomacro_validate_json_number (data -> 'Duree');
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

ALTER TABLE camps
    ADD CONSTRAINT Navette_gomacro CHECK (gomacro_validate_json_camp_Navette (Navette));

ALTER TABLE camps
    ADD CONSTRAINT OptionPrix_gomacro CHECK (gomacro_validate_json_camp_OptionPrixCamp (OptionPrix));

ALTER TABLE participants
    ADD CONSTRAINT OptionPrix_gomacro CHECK (gomacro_validate_json_camp_OptionPrixParticipant (OptionPrix));

ALTER TABLE equipiers
    ADD CONSTRAINT Presence_gomacro CHECK (gomacro_validate_json_camp_OptionnalPlage (Presence));

ALTER TABLE participants
    ADD CONSTRAINT Remises_gomacro CHECK (gomacro_validate_json_camp_Remises (Remises));

ALTER TABLE groupes
    ADD CONSTRAINT Plage_gomacro CHECK (gomacro_validate_json_shar_Plage (Plage));

