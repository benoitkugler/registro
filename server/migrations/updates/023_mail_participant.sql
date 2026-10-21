-- v0.12.3
-- ajout champ Mail sur formulaire d'inscription
BEGIN;

-- 
ALTER TABLE
    inscription_participants
ADD
    COLUMN Mail text;

ALTER TABLE
    inscription_participants
ADD
    COLUMN IsDoublon2 boolean;

UPDATE
    inscription_participants
SET
    Mail = '';

UPDATE
    inscription_participants
SET
    isDoublon2 = isDoublon;

ALTER TABLE
    inscription_participants DROP COLUMN isDoublon;

ALTER TABLE
    inscription_participants RENAME COLUMN isDoublon2 TO isDoublon;

ALTER TABLE
    inscription_participants
ALTER COLUMN
    Mail
SET
    NOT NULL;

ALTER TABLE
    inscription_participants
ALTER COLUMN
    isDoublon
SET
    NOT NULL;

COMMIT;