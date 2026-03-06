-- v0.10.3
-- properly add cascade on foreign key

BEGIN;
ALTER TABLE ficheequipiers
    DROP CONSTRAINT ficheequipiers_idpersonne_guard_fkey;
ALTER TABLE ficheequipiers
    ADD FOREIGN KEY (IdPersonne, guard) REFERENCES personnes (Id, IsTemp) ON DELETE CASCADE;
ALTER TABLE fichesanitaires
    DROP CONSTRAINT fichesanitaires_idpersonne_guard_fkey;
ALTER TABLE fichesanitaires
    ADD FOREIGN KEY (IdPersonne, guard) REFERENCES personnes (Id, IsTemp) ON DELETE CASCADE;
COMMIT;

