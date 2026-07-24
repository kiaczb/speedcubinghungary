-- A report_week visszavonható; a post_tag enum 'report' értéke NEM (a Postgres
-- nem támogatja az enum érték eltávolítását). Ez tudatos aszimmetria: a lefelé
-- migráció a séma használható állapotát állítja vissza, az árva enum érték
-- ártalmatlan.
ALTER TABLE posts DROP COLUMN report_week;
