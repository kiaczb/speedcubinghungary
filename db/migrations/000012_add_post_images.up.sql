-- Egy poszt több képet is hordozhat (galéria / carousel a poszt oldalán).
-- Tömb és nem külön tábla, mert a SORREND maga az adat (az első elem a
-- borítókép, ami a listákban és a NewsCard-on látszik), így nem kell külön
-- ordinal oszlop és join. A képek száma alkalmazásszinten korlátozott
-- (post.MaxImages = 10), a tömb az útvonalakat tárolja, nem a fájlokat.
ALTER TABLE posts ADD COLUMN image_paths TEXT[] NOT NULL DEFAULT '{}';

-- A meglévő egyetlen borítóképből egyelemű galéria lesz.
UPDATE posts
SET image_paths = ARRAY[image_path]
WHERE image_path IS NOT NULL AND image_path <> '';

ALTER TABLE posts DROP COLUMN image_path;
