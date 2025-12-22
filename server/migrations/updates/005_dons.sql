-- for v0.8.1
BEGIN;
ALTER TABLE dons
    DROP COLUMN IdPaiementHelloasso;
ALTER TABLE dons
    ADD COLUMN IdPaiementHelloasso integer NOT NULL;
COMMIT;

