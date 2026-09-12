-- v0.12
-- permet de restreindre les messages à un seul directeur
BEGIN;

ALTER TABLE
    event_messages DROP COLUMN guard;

ALTER TABLE
    event_messages
ADD
    COLUMN OnlyToCamp int;

ALTER TABLE
    event_messages
ADD
    COLUMN guard smallint;

UPDATE
    event_messages
SET
    guard = 2;

ALTER TABLE
    event_messages
ALTER COLUMN
    guard
SET
    NOT NULL;

ALTER TABLE
    event_messages
ADD
    FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE
    event_messages
ALTER COLUMN
    guard
SET
    DEFAULT 2
    /* EventKind.Message */
;

ALTER TABLE
    event_messages
ADD
    CHECK (
        guard = 2
        /* EventKind.Message */
    );

ALTER TABLE
    event_messages
ADD
    CHECK (
        OnlyToFondSoutien = FALSE
        OR OnlyToCamp IS NULL
    );

ALTER TABLE
    event_messages
ADD
    CHECK (
        (
            OnlyToFondSoutien = FALSE
            AND OnlyToCamp IS NULL
        )
        OR Origine = 0
        /* Acteur.Espaceperso */
    );

ALTER TABLE
    event_messages
ADD
    FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE
    event_messages
ADD
    FOREIGN KEY (OrigineCamp) REFERENCES camps;

ALTER TABLE
    event_messages
ADD
    FOREIGN KEY (OnlyToCamp) REFERENCES camps;

COMMIT;