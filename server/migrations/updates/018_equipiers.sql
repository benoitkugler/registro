-- v0.12 
BEGIN;

ALTER TABLE
    ficheequipiers DROP CONSTRAINT ficheequipiers_diplome_check;

ALTER TABLE
    ficheequipiers
ADD
    CONSTRAINT ficheequipiers_diplome_check CHECK (
        Diplome IN (
            0,
            1,
            2,
            3,
            4,
            5,
            6,
            7,
            8,
            9,
            10,
            11,
            12,
            13,
            14,
            15,
            16,
            17,
            18,
            19,
            20,
            21,
            22,
            23,
            24
        )
    );

ALTER TABLE
    ficheequipiers DROP COLUMN guard;

ALTER TABLE
    ficheequipiers
ADD
    COLUMN FormationRepere smallint;

UPDATE
    ficheequipiers
SET
    FormationRepere = 0;

ALTER TABLE
    ficheequipiers
ALTER COLUMN
    FormationRepere
SET
    NOT NULL;

ALTER TABLE
    ficheequipiers
ADD
    CONSTRAINT ficheequipiers_FormationRepere_check CHECK (FormationRepere IN (0, 1, 2, 3, 4));

ALTER TABLE
    ficheequipiers
ADD
    COLUMN guard boolean;

UPDATE
    ficheequipiers
SET
    guard = FALSE;

ALTER TABLE
    ficheequipiers
ALTER COLUMN
    guard
SET
    NOT NULL;

ALTER TABLE
    ficheequipiers
ADD
    FOREIGN KEY (IdPersonne, guard) REFERENCES personnes (Id, IsTemp) ON DELETE CASCADE;

ALTER TABLE
    ficheequipiers
ALTER COLUMN
    guard
SET
    DEFAULT FALSE;

ALTER TABLE
    ficheequipiers
ADD
    CHECK (guard = FALSE);

COMMIT;