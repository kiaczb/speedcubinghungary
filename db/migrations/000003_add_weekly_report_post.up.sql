-- A heti riport a MEGLÉVŐ posts rendszerbe kerül, nem külön táblába:
-- így örökli a listázást, a részletező oldalt és a jogosultságkezelést.

-- 1) Külön kategória a riportoknak, hogy megkülönböztethetők/kereshetők legyenek.
--    PG12+ óta az ADD VALUE tranzakcióban is futhat (az új érték csak a következő
--    tranzakcióban használható — ez a migráció nem is használja).
ALTER TYPE post_tag ADD VALUE IF NOT EXISTS 'report';

-- 2) Idempotencia-kulcs: a riportolt hét hétfője (ISO hét kezdete).
--    UNIQUE, de a Postgres a NULL-okat külön értéknek veszi, ezért a normál
--    posztok (report_week IS NULL) korlátozás nélkül maradnak — hetente viszont
--    csak egyetlen riport jöhet létre (ON CONFLICT DO NOTHING a job oldalán).
ALTER TABLE posts ADD COLUMN report_week DATE UNIQUE;
