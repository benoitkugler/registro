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
        (1, 'Carte d''identité/Passeport', 2, 0),
        --
        (2, 'Permis de conduire', 2, 0),
        --
        (3, 'Surveillant de baignade', 1, 0),
        --
        (4, 'Secourisme (PSC1 - AFPS)', 1, 0),
        --
        (5, 'BAFA', 1, 0),
        --
        (6, 'BAFD', 1, 0),
        --
        (7, 'Carte Vitale', 2, 0),
        --
        (8, 'Vaccin', 5, 0),
        --
        (9, 'Certificat concernant les normes HACCP, délivré sur place après une formation.', 1, 0),
        --
        (10, 'Equivalent BAFD', 1, 0),
        --
        (11, 'Equivalent BAFA', 1, 0),
        --
        (12, 'Certificat de non contre-indication à la cuisine de collectivité, à demander à son médecin traitant.', 1, 90),
        --
        (13, 'Certificat de scolarité', 1, 0),
        --
        (14, 'Autre', 1, 0);

SELECT
    setval('demandes_id_seq', (
            SELECT
                max(id)
            FROM demandes));

