-- Visszafelé csak a borítókép (az első elem) őrizhető meg; a többi kép fájlja
-- a volume-on marad, de már nem hivatkozott.
ALTER TABLE posts ADD COLUMN image_path TEXT;

UPDATE posts SET image_path = image_paths[1];

ALTER TABLE posts DROP COLUMN image_paths;
