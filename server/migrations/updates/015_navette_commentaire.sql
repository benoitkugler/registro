-- v0.12 

BEGIN;
ALTER TABLE participants DROP COLUMN Navette;
ALTER TABLE participants DROP COLUMN Commentaire;
COMMIT;
