--
-- Script to be run after the DB creation (mandatory, assumed by the server code)
--
--
-- Default Taux

INSERT INTO tauxs
    VALUES (1, 'Euros seulement (par défaut)', 1000, 0);

SELECT
    setval('tauxs_id_seq', (
            SELECT
                max(id)
            FROM tauxs));

-- Builtin Demandes
INSERT INTO demandes (Categorie, Description, MaxDocs, JoursValide)
    VALUES
        --
        (1, '', 2, 0),
        --
        (2, '', 2, 0),
        --
        (3, '', 1, 0),
        --
        (4, '', 1, 0),
        --
        (5, '', 1, 0),
        --
        (6, '', 1, 0),
        --
        (7, '', 2, 0),
        --
        (8, 'Vaccins', 5, 0),
        --
        (9, 'Certificat délivré sur place après une formation.', 1, 0),
        --
        (10, '', 1, 0),
        --
        (11, '', 1, 0),
        --
        (12, 'Non contre-indication à la cuisine de collectivité, à demander à son médecin traitant.', 1, 90),
        --
        (13, '', 1, 0),
        --
        (14, '', 1, 0),
        --
        (15, '', 1, 0);

SELECT
    setval('demandes_id_seq', (
            SELECT
                max(id)
            FROM demandes));

