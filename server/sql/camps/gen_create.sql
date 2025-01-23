-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE camps (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    DateDebut date NOT NULL,
    Duree integer NOT NULL,
    Agrement text NOT NULL
);

CREATE TABLE participants (
    Id serial PRIMARY KEY,
    IdCamp integer NOT NULL,
    IdPersonne integer NOT NULL
);

-- constraints
ALTER TABLE participants
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE participants
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes;

