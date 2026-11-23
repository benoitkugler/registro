-- v0.13.1
-- La clé V1 est facultative : plusieurs dossiers peuvent avoir une clé vide.
-- L'unicité ne s'applique qu'aux clés non vides.

BEGIN;

ALTER TABLE dossiers DROP CONSTRAINT IF EXISTS dossiers_keyv1_key;

CREATE UNIQUE INDEX ON dossiers (KeyV1)
WHERE
    length(KeyV1) > 0;

COMMIT;
