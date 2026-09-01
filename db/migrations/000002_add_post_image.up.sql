-- Optional cover image for a post. Stores the public path served by the API
-- (e.g. '/uploads/<uuid>.webp'), not the file bytes — the file itself lives on
-- the uploads volume mounted into the api container.
ALTER TABLE posts ADD COLUMN image_path TEXT;
