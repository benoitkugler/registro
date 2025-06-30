-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE events (
    Id serial PRIMARY KEY,
    IdDossier integer NOT NULL,
    Kind smallint CHECK (Kind IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL,
    Created timestamp(0) with time zone NOT NULL
);

CREATE TABLE event_attestations (
    IdEvent integer NOT NULL,
    Distribution smallint CHECK (Distribution IN (0, 1, 2)) NOT NULL,
    IsPresence boolean NOT NULL,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

CREATE TABLE event_camp_docss (
    IdEvent integer NOT NULL,
    IdCamp integer NOT NULL,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

CREATE TABLE event_messages (
    IdEvent integer NOT NULL,
    Contenu text NOT NULL,
    Origine smallint CHECK (Origine IN (0, 1, 2)) NOT NULL,
    OrigineCamp integer,
    VuBackoffice boolean NOT NULL,
    VuEspaceperso boolean NOT NULL,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

CREATE TABLE event_message_vus (
    IdEvent integer NOT NULL,
    IdCamp integer NOT NULL,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

CREATE TABLE event_place_liberees (
    IdEvent integer NOT NULL,
    IdParticipant integer NOT NULL,
    Accepted boolean NOT NULL,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

CREATE TABLE event_sondages (
    IdEvent integer NOT NULL,
    IdCamp integer NOT NULL,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

CREATE TABLE event_validations (
    IdEvent integer NOT NULL,
    IdCamp integer,
    guard smallint CHECK (guard IN (0, 1, 2, 3, 4, 5, 6, 7)) NOT NULL
);

-- constraints
ALTER TABLE events
    ADD UNIQUE (Id, Kind);

ALTER TABLE events
    ADD FOREIGN KEY (IdDossier) REFERENCES dossiers ON DELETE CASCADE;

ALTER TABLE event_validations
    ADD UNIQUE (IdEvent);

ALTER TABLE event_validations
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_validations
    ADD FOREIGN KEY (IdEvent) REFERENCES events;

ALTER TABLE event_validations
    ALTER COLUMN guard SET DEFAULT 1
    /* EventKind.Validation */
;

ALTER TABLE event_validations
    ADD CHECK (guard = 1
    /* EventKind.Validation */);

ALTER TABLE event_messages
    ADD UNIQUE (IdEvent);

ALTER TABLE event_messages
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_messages
    ADD CHECK (Origine <> 2
    /* MessageOrigine.FromDirecteur */
        OR OrigineCamp IS NOT NULL);

ALTER TABLE event_messages
    ADD CHECK (Origine = 2
    /* MessageOrigine.FromDirecteur */
        OR OrigineCamp IS NULL);

ALTER TABLE event_messages
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE event_messages
    ADD FOREIGN KEY (OrigineCamp) REFERENCES camps;

ALTER TABLE event_messages
    ALTER COLUMN guard SET DEFAULT 2
    /* EventKind.Message */
;

ALTER TABLE event_messages
    ADD CHECK (guard = 2
    /* EventKind.Message */);

ALTER TABLE event_message_vus
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_message_vus
    ADD UNIQUE (IdEvent, IdCamp);

ALTER TABLE event_message_vus
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE event_message_vus
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;

ALTER TABLE event_message_vus
    ALTER COLUMN guard SET DEFAULT 2
    /* EventKind.Message */
;

ALTER TABLE event_message_vus
    ADD CHECK (guard = 2
    /* EventKind.Message */);

ALTER TABLE event_camp_docss
    ADD UNIQUE (IdEvent);

ALTER TABLE event_camp_docss
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_camp_docss
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE event_camp_docss
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE event_camp_docss
    ALTER COLUMN guard SET DEFAULT 5
    /* EventKind.CampDocs */
;

ALTER TABLE event_camp_docss
    ADD CHECK (guard = 5
    /* EventKind.CampDocs */);

ALTER TABLE event_sondages
    ADD UNIQUE (IdEvent);

ALTER TABLE event_sondages
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_sondages
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE event_sondages
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

ALTER TABLE event_sondages
    ALTER COLUMN guard SET DEFAULT 7
    /* EventKind.Sondage */
;

ALTER TABLE event_sondages
    ADD CHECK (guard = 7
    /* EventKind.Sondage */);

ALTER TABLE event_place_liberees
    ADD UNIQUE (IdEvent);

ALTER TABLE event_place_liberees
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_place_liberees
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE event_place_liberees
    ADD FOREIGN KEY (IdParticipant) REFERENCES participants;

ALTER TABLE event_place_liberees
    ALTER COLUMN guard SET DEFAULT 3
    /* EventKind.PlaceLiberee */
;

ALTER TABLE event_place_liberees
    ADD CHECK (guard = 3
    /* EventKind.PlaceLiberee */);

ALTER TABLE event_attestations
    ADD UNIQUE (IdEvent);

ALTER TABLE event_attestations
    ADD FOREIGN KEY (IdEvent, guard) REFERENCES events (Id, Kind) ON DELETE CASCADE;

ALTER TABLE event_attestations
    ADD FOREIGN KEY (IdEvent) REFERENCES events ON DELETE CASCADE;

ALTER TABLE event_attestations
    ALTER COLUMN guard SET DEFAULT 6
    /* EventKind.Attestation */
;

ALTER TABLE event_attestations
    ADD CHECK (guard = 6
    /* EventKind.Attestation */);

