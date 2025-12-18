BEGIN;
ALTER TABLE groupes
    DROP COLUMN plage;
ALTER TABLE groupes
    DROP COLUMN couleur;
ALTER TABLE groupes
    ADD COLUMN Couleur text NOT NULL;
ALTER TABLE groupes
    ADD COLUMN Fin date NOT NULL;
COMMIT;

