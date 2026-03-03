BEGIN;
--
CREATE TABLE projet_spis (
    IdCamp integer NOT NULL,
    Description text NOT NULL,
    Programme text NOT NULL,
    JourneeType text NOT NULL,
    DynamiqueCampeur text NOT NULL,
    Evangile text NOT NULL,
    Equipe text NOT NULL,
    Cuisine text NOT NULL,
    Suite text NOT NULL,
    VisiteLibrairie smallint CHECK (VisiteLibrairie IN (0, 1, 2)) NOT NULL,
    Bibles boolean NOT NULL,
    Question text NOT NULL
);
-- constraints
ALTER TABLE projet_spis
    ADD UNIQUE (IdCamp);
ALTER TABLE projet_spis
    ADD FOREIGN KEY (IdCamp) REFERENCES camps ON DELETE CASCADE;
--
--

COMMIT;

