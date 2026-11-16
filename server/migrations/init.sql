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
INSERT INTO demandes (Categorie, Description, MaxDocs)
    VALUES
        --
        (1, '', 2),
        --
        (2, '', 2),
        --
        (3, '', 1),
        --
        (4, '', 1),
        --
        (5, '', 1),
        --
        (6, '', 1),
        --
        (7, '', 2),
        --
        (8, 'Vaccins', 5),
        --
        (9, '', 1),
        --
        (10, '', 1),
        --
        (11, '', 1);

SELECT
    setval('demandes_id_seq', (
            SELECT
                max(id)
            FROM demandes));

-- default compta config

INSERT INTO compta_configs (Year, -- integer
    MealsCost, -- integer[]
    SmallMealThreshold, -- integer
    SmallMealMajoration, -- integer
    AssoCars, -- integer[]
    ChauffageMDR, -- integer
    EntretienMDR, -- integer
    MaterielMDR -- integer
 )
    VALUES (2020, '{50, 200, 50, 200, 70}', 25, 10, '{150, 80, 50}', 15000, 20, 100);

