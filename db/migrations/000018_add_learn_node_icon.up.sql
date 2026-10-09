-- Opcionális ikon egy Tanulás-elemhez: a @cubing/icons egy osztályneve
-- (pl. 'event-333', 'unofficial-fto'), üres = nincs ikon. Adat, nem kód: az
-- admin választja ki, így a kirakók továbbra sincsenek a kódban felsorolva.
-- A feltöltött kép (ha megépül) külön oszlop lesz, és elsőbbséget kap.
ALTER TABLE learn_nodes ADD COLUMN icon TEXT NOT NULL DEFAULT '';
