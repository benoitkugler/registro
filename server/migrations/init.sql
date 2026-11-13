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
        (9, '', 1, 0),
        --
        (10, '', 1, 0),
        --
        (11, '', 1, 0);

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

