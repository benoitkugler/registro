-- v0.12
-- Ajoute un event lors d'un changement de camp
BEGIN;

ALTER TABLE
    events DROP CONSTRAINT events_kind_check;

ALTER TABLE
    events
ADD
    CONSTRAINT events_kind_check CHECK (Kind IN (0, 1, 2, 3, 4, 5, 6, 7, 8));

CREATE TABLE event_changement_camps (
    IdEvent integer NOT NULL,
    IdParticipant integer NOT NULL,
    Old integer NOT NULL,
    New integer NOT NULL,
    guard smallint NOT NULL
);

ALTER TABLE
    event_changement_camps
ADD
    UNIQUE (IdEvent);

ALTER TABLE
    event_changement_camps
ADD
    FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE
    event_changement_camps
ADD
    FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE
    event_changement_camps
ADD
    FOREIGN KEY (IdParticipant) REFERENCES participants;

ALTER TABLE
    event_changement_camps
ADD
    FOREIGN KEY (Old) REFERENCES camps;

ALTER TABLE
    event_changement_camps
ADD
    FOREIGN KEY (New) REFERENCES camps;

ALTER TABLE
    event_changement_camps
ALTER COLUMN
    guard
SET
    DEFAULT 8
    /* EventKind.ChangementCamp */
;

ALTER TABLE
    event_changement_camps
ADD
    CHECK (
        guard = 8
        /* EventKind.ChangementCamp */
    );

COMMIT;