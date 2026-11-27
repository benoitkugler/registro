-- Attestation de quotient familial : un fichier par dossier.
-- Le fichier est téléversé depuis l'espace personnel, puis consulté
-- et validé (ou refusé) depuis le backoffice.

BEGIN;

CREATE TABLE file_quotient_familials (
    IdFile integer NOT NULL,
    IdDossier integer NOT NULL
);

ALTER TABLE file_quotient_familials
    ADD UNIQUE (IdFile);

ALTER TABLE file_quotient_familials
    ADD UNIQUE (IdDossier);

ALTER TABLE file_quotient_familials
    ADD FOREIGN KEY (IdFile) REFERENCES files ON DELETE CASCADE;

ALTER TABLE file_quotient_familials
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE RESTRICT;

COMMIT;
