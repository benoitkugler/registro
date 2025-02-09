-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TYPE Montant AS (
    Cent integer,
    Currency smallint
);

CREATE TABLE dossiers (
    Id serial PRIMARY KEY,
    IdTaux integer NOT NULL,
    IdResponsable integer NOT NULL,
    CopiesMails text[],
    PartageAdressesOK boolean NOT NULL,
    IsValidated boolean NOT NULL,
    LastConnection timestamp(0) with time zone NOT NULL,
    KeyV1 text NOT NULL
);

CREATE TABLE events (
    Id serial PRIMARY KEY,
    IdDossier integer NOT NULL,
    Kind smallint CHECK (Kind IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9)) NOT NULL,
    Created timestamp(0) with time zone NOT NULL
);

CREATE TABLE event_messages (
    IdEvent integer NOT NULL,
    Guard smallint CHECK (Guard IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9)) NOT NULL,
    Contenu text NOT NULL,
    Origine smallint CHECK (Origine IN (0, 1, 2)) NOT NULL,
    VuPar boolean[] CHECK (array_length(VuPar, 1) = 3) NOT NULL
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

CREATE TABLE tauxs (
    Id serial PRIMARY KEY,
    Label text NOT NULL,
    Euros integer NOT NULL,
    FrancsSuisse integer NOT NULL
);

-- constraints
ALTER TABLE tauxs
    ADD UNIQUE (Label);

ALTER TABLE tauxs
    ADD CHECK (Euros = 1000);

ALTER TABLE dossiers
    ADD UNIQUE (Id, IdTaux);

ALTER TABLE dossiers
    ADD FOREIGN KEY (IdTaux) REFERENCES tauxs;

ALTER TABLE dossiers
    ADD FOREIGN KEY (IdResponsable) REFERENCES personnes;

ALTER TABLE paiements
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE events
    ADD UNIQUE (Id, Kind);

ALTER TABLE events
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE event_messages
    ADD UNIQUE (IdEvent);

ALTER TABLE event_messages
    ADD CHECK (guard = 1
    /* EventKind.Message */);

ALTER TABLE event_messages
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (id, kind);

ALTER TABLE event_messages
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

