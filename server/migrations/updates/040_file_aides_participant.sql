-- v0.13.2
-- Un justificatif d'aide peut être téléversé par les parents avant la création de l'aide.
-- Le fichier est alors rattaché au participant (IdAide NULL), puis à l'aide lors de sa création.
-- Les anciennes aides non validées redeviennent de simples justificatifs en attente :
-- le champ Aide.Valide n'a plus d'utilité.
-- On considère qu'il existe au plus une aide par participant.

BEGIN;

-- Nécessaire pour la clé étrangère composite de file_aides.
ALTER TABLE aides
    ADD UNIQUE (Id, IdParticipant);

-- Au plus une aide par participant.
ALTER TABLE aides
    ADD UNIQUE (IdParticipant);


ALTER TABLE file_aides
    ADD COLUMN IdParticipant integer;
    
ALTER TABLE file_aides
    ADD UNIQUE (IdParticipant);

UPDATE file_aides
    SET IdParticipant = aides.IdParticipant
FROM aides
WHERE file_aides.IdAide = aides.Id;

ALTER TABLE file_aides
    ALTER COLUMN IdParticipant SET NOT NULL;

ALTER TABLE file_aides
    ALTER COLUMN IdAide DROP NOT NULL;

UPDATE file_aides
    SET IdAide = NULL
FROM aides
WHERE file_aides.IdAide = aides.Id AND NOT aides.Valide;

DELETE FROM aides
    WHERE NOT Valide;

ALTER TABLE aides
    DROP COLUMN Valide;

ALTER TABLE file_aides
    DROP CONSTRAINT file_aides_idaide_fkey;

ALTER TABLE file_aides
    ADD FOREIGN KEY (IdAide, IdParticipant) REFERENCES aides (Id, IdParticipant);

ALTER TABLE file_aides
    ADD FOREIGN KEY (IdParticipant) REFERENCES participants;

ALTER TABLE file_aides
    ADD FOREIGN KEY (IdAide) REFERENCES aides ON DELETE RESTRICT;

COMMIT;
