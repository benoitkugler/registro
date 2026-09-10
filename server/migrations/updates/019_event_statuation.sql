-- v0.12
-- best effort to update the statuation events,
-- for each dossier which has at least one validation event
BEGIN;

CREATE TABLE __migration_19 (
    id SERIAL,
    iddossier int,
    created timestamp(0) with time zone,
    idcamp int,
    isbackoffice boolean,
    idParticipant int,
    statut int
);

WITH validations AS (
    SELECT
        dossiers.id as iddossier,
        (
            SELECT
                events.id
            FROM
                events
            WHERE
                events.iddossier = dossiers.id
                AND events.kind = 1
            ORDER BY
                events.created DESC
            LIMIT
                1
        ) as idevent
    FROM
        dossiers
)
INSERT INTO
    __migration_19 (
        iddossier,
        created,
        idcamp,
        isbackoffice,
        idParticipant,
        statut
    )
SELECT
    validations.iddossier,
    events.created,
    event_validations.idcamp,
    event_validations.isbackoffice,
    participants.id as idParticipant,
    participants.statut
FROM
    validations
    JOIN events ON events.id = validations.idevent
    JOIN event_validations ON event_validations.idevent = validations.idevent
    JOIN participants ON participants.iddossier = validations.iddossier
WHERE
    validations.idevent IS NOT NULL;

-- delete all existing Validation event and the table
-- now we have saved them in  __migration_19
DELETE FROM
    events
WHERE
    Kind = 1;

-- empty now, by cascade
DROP TABLE event_validations;

-- create the new table 
CREATE TABLE event_statuations (
    IdEvent integer NOT NULL,
    IdCamp integer NOT NULL,
    IsBackoffice boolean NOT NULL,
    IdParticipant integer NOT NULL,
    Statut smallint CHECK (Statut IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    guard smallint NOT NULL
);

ALTER TABLE
    event_statuations
ADD
    UNIQUE (IdEvent);

ALTER TABLE
    event_statuations
ADD
    FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE
    event_statuations
ADD
    FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE
    event_statuations
ADD
    FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE
    event_statuations
ADD
    FOREIGN KEY (IdParticipant) REFERENCES participants ON DELETE CASCADE;

ALTER TABLE
    event_statuations
ALTER COLUMN
    guard
SET
    DEFAULT 1
    /* EventKind.Statuation */
;

ALTER TABLE
    event_statuations
ADD
    CHECK (
        guard = 1
        /* EventKind.Statuation */
    );

-- create the new events, registering the IdMigration key 
ALTER TABLE
    events
ADD
    COLUMN IdMigration int;

INSERT INTO
    events (IdDossier, Kind, Created, IdMigration)
SELECT
    iddossier,
    1,
    created,
    id
FROM
    __migration_19;

-- create the Statuation events
INSERT INTO
    event_statuations (
        IdEvent,
        IdCamp,
        IsBackoffice,
        IdParticipant,
        Statut
    )
SELECT
    events.id,
    __migration_19.idcamp,
    __migration_19.isbackoffice,
    __migration_19.idParticipant,
    __migration_19.statut
FROM
    __migration_19
    JOIN events on events.IdMigration = __migration_19.Id;

ALTER TABLE
    events DROP COLUMN IdMigration;

DROP TABLE __migration_19;

COMMIT;