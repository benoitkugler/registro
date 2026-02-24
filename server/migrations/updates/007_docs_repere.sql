-- add two new categories
BEGIN;
ALTER TABLE demandes
    DROP CONSTRAINT demandes_categorie_check;
ALTER TABLE demandes
    ADD CONSTRAINT demandes_categorie_check CHECK (Categorie IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15));
INSERT INTO demandes (Categorie, Description, MaxDocs, JoursValide)
    VALUES
        --
        (14, '', 1, 0),
        --
        (15, '', 1, 0);
SELECT
    setval('demandes_id_seq', (
            SELECT
                max(id)
            FROM demandes));
COMMIT;

