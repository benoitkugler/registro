-- v0.13.1
-- Add a flag on camps to ask for a deposit (acompte) on validated inscriptions.
-- The deposit amount is computed as a percentage of the camp price,
-- see config.Asso.AcomptePercent.

BEGIN;
ALTER TABLE camps ADD COLUMN DemandeAcompte boolean NOT NULL DEFAULT false;
ALTER TABLE camps ALTER COLUMN DemandeAcompte DROP DEFAULT;
COMMIT;
