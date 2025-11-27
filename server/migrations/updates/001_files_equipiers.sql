-- Remove the Scolarite categorie; assume no files with Scolarite or CertMedicalCuisine
-- have been uploaded

DO $$
DECLARE
    files_count integer;
BEGIN
    SELECT
        count(IdFile) INTO files_count
    FROM
        file_personnes
    WHERE
        iddemande = 13;
    --
    assert files_count = 0,
    'Files with categorie Scolarite exist !';
END
$$;

--
--

BEGIN;
--
DELETE FROM demande_equipiers
WHERE iddemande = 13;
-- map 14 to 13
UPDATE
    demande_equipiers
SET
    iddemande = 13
WHERE
    iddemande = 14;
--
UPDATE
    file_personnes
SET
    iddemande = 13
WHERE
    iddemande = 14;
--
DELETE FROM Demandes
WHERE id = 14;
--
SELECT
    setval('demandes_id_seq', (
            SELECT
                max(id)
            FROM demandes));
COMMIT;

