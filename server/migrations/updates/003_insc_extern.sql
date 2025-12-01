BEGIN;
ALTER TABLE camps RENAME COLUMN WithoutInscription TO InscriptionExterne;
COMMIT;

