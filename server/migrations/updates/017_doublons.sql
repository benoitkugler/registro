-- v0.12 
BEGIN;

-- create a new table with the proper fields 
ALTER TABLE
    inscription_participants RENAME TO inscription_participants_old;

CREATE TABLE inscription_participants (
    Id serial PRIMARY KEY,
    IdInscription integer NOT NULL,
    IdCamp integer NOT NULL,
    IdTaux integer NOT NULL,
    Nom text NOT NULL,
    Prenom text NOT NULL,
    DateNaissance date NOT NULL,
    Sexe smallint CHECK (Sexe IN (0, 1, 2)) NOT NULL,
    Nationnalite Nationnalite NOT NULL,
    IsDoublon boolean NOT NULL
);

ALTER TABLE
    inscription_participants
ADD
    FOREIGN KEY (IdCamp, IdTaux) REFERENCES camps (Id, IdTaux) ON DELETE CASCADE;

ALTER TABLE
    inscription_participants
ADD
    FOREIGN KEY (IdInscription, IdTaux) REFERENCES inscriptions (Id, IdTaux) ON DELETE CASCADE;

ALTER TABLE
    inscription_participants
ADD
    FOREIGN KEY (IdInscription) REFERENCES inscriptions ON DELETE CASCADE;

ALTER TABLE
    inscription_participants
ADD
    FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE
    inscription_participants
ADD
    FOREIGN KEY (IdTaux) REFERENCES tauxs;

INSERT INTO
    inscription_participants (
        IdInscription,
        IdCamp,
        IdTaux,
        Nom,
        Prenom,
        DateNaissance,
        Sexe,
        Nationnalite,
        IsDoublon
    )
SELECT
    IdInscription,
    IdCamp,
    IdTaux,
    Nom,
    Prenom,
    DateNaissance,
    Sexe,
    Nationnalite,
    false
FROM
    inscription_participants_old;

DROP TABLE inscription_participants_old;

COMMIT;