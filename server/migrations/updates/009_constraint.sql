-- v0.10.1
-- fix missing constraint

ALTER TABLE event_validations
    ADD FOREIGN KEY (IdCamp) REFERENCES camps;

