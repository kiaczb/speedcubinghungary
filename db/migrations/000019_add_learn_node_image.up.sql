-- A kártya képe: ma egy https:// kép-URL (pl. VisualCube-ábra), később ide
-- kerül a feltöltött kép útvonala is. Üres = nincs kép; a felület ilyenkor az
-- ikont (000018), ha az sincs, a cím kezdőbetűjét mutatja.
ALTER TABLE learn_nodes ADD COLUMN image_url TEXT NOT NULL DEFAULT '';
