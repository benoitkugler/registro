-- v0.11.3
-- add a boolean field to Camp

BEGIN;
ALTER TABLE camps ADD COLUMN IsAlbumVisible boolean;
-- set to true if there is an album 
UPDATE camps SET IsAlbumVisible = (AlbumID <> '');
ALTER TABLE camps ALTER COLUMN IsAlbumVisible SET NOT NULL;
COMMIT;

