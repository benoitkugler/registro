-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE demandes (
    Id serial PRIMARY KEY,
    IdFile integer,
    IdDirecteur integer,
    Categorie smallint CHECK (Categorie IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)) NOT NULL,
    Description text NOT NULL,
    MaxDocs integer NOT NULL,
    JoursValide integer NOT NULL
);

CREATE TABLE demande_camps (
    IdCamp integer NOT NULL,
    IdDemande integer NOT NULL
);

CREATE TABLE demande_equipiers (
    IdEquipier integer NOT NULL,
    IdDemande integer NOT NULL,
    Optionnel boolean NOT NULL
);

CREATE TABLE files (
    Id serial PRIMARY KEY,
    Taille integer NOT NULL,
    NomClient text NOT NULL,
    DateHeureModif timestamp(0) with time zone NOT NULL
);

CREATE TABLE file_aides (
    IdFile integer NOT NULL,
    IdAide integer NOT NULL
);

CREATE TABLE file_camps (
    IdFile integer NOT NULL,
    IdCamp integer NOT NULL,
    IsLettre boolean NOT NULL
);

CREATE TABLE file_personnes (
    IdFile integer NOT NULL,
    IdPersonne integer NOT NULL,
    IdDemande integer NOT NULL
);

-- constraints
ALTER TABLE demandes
    ADD CONSTRAINT constraint_categorie CHECK (Categorie = 0 OR IdDirecteur IS NULL);

ALTER TABLE demandes
    ADD CONSTRAINT constraint_maxdocs CHECK (MaxDocs >= 1);

CREATE UNIQUE INDEX ON demandes (Categorie)
WHERE
    Categorie <> 0;

ALTER TABLE demandes
    ADD FOREIGN KEY (IdFile) REFERENCES files;

ALTER TABLE demandes
    ADD FOREIGN KEY (IdDirecteur) REFERENCES personnes ON DELETE CASCADE;

ALTER TABLE demande_equipiers
    ADD UNIQUE (IdEquipier, IdDemande);

ALTER TABLE demande_equipiers
    ADD FOREIGN KEY (IdEquipier) REFERENCES equipiers ON DELETE CASCADE;

ALTER TABLE demande_equipiers
    ADD FOREIGN KEY (IdDemande) REFERENCES demandes;

ALTER TABLE demande_camps
    ADD UNIQUE (IdCamp, IdDemande);

ALTER TABLE demande_camps
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE demande_camps
    ADD FOREIGN KEY (IdDemande) REFERENCES demandes;

ALTER TABLE file_camps
    ADD UNIQUE (IdFile);

CREATE UNIQUE INDEX ON file_camps (IdCamp)
WHERE
    IsLettre IS TRUE;

ALTER TABLE file_camps
    ADD FOREIGN KEY (IdFile) REFERENCES files ON DELETE CASCADE;

ALTER TABLE file_camps
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE file_personnes
    ADD UNIQUE (IdFile);

ALTER TABLE file_personnes
    ADD FOREIGN KEY (IdFile) REFERENCES files ON DELETE CASCADE;

ALTER TABLE file_personnes
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes;

ALTER TABLE file_personnes
    ADD FOREIGN KEY (IdDemande) REFERENCES demandes;

ALTER TABLE file_aides
    ADD UNIQUE (IdFile);

ALTER TABLE file_aides
    ADD UNIQUE (IdAide);

ALTER TABLE file_aides
    ADD FOREIGN KEY (IdFile) REFERENCES files ON DELETE CASCADE;

ALTER TABLE file_aides
    ADD FOREIGN KEY (IdAide) REFERENCES aides;

