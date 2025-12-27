-- for v0.8.1
BEGIN;
DROP TABLE dons CASCADE;
CREATE TABLE dons (
    Id serial PRIMARY KEY,
    IdPersonne integer,
    IdOrganisme integer,
    Montant Montant NOT NULL,
    ModePaiement smallint CHECK (ModePaiement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    date date NOT NULL,
    Affectation text NOT NULL,
    Details text NOT NULL,
    Remercie boolean NOT NULL
);
CREATE TABLE organismes (
    Id serial PRIMARY KEY,
    Nom text NOT NULL,
    Mail text NOT NULL,
    Adresse text NOT NULL,
    CodePostal text NOT NULL,
    Ville text NOT NULL,
    Pays text NOT NULL
);
ALTER TABLE dons
    ADD FOREIGN KEY (IdPersonne) REFERENCES personnes;
ALTER TABLE dons
    ADD FOREIGN KEY (IdOrganisme) REFERENCES organismes;
COMMIT;

