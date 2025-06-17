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
    DemandeFondSoutien boolean NOT NULL,
    IsValidated boolean NOT NULL,
    MomentInscription timestamp(0) with time zone NOT NULL,
    LastLoadDocuments timestamp(0) with time zone NOT NULL,
    KeyV1 text NOT NULL
);

CREATE TABLE paiements (
    Id serial PRIMARY KEY,
    IdDossier integer NOT NULL,
    IsRemboursement boolean NOT NULL,
    Montant Montant NOT NULL,
    Payeur text NOT NULL,
    Mode smallint CHECK (Mode IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Time timestamp(0) with time zone NOT NULL,
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

