-- A megszorítás eldobható; az 1) lépésben átnevezett sorok viszont NEM
-- állíthatók vissza (nem tudnánk, melyik volt eredetileg üres) — ez szándékos,
-- azok a sorok hibás adatok voltak.
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_wca_id_not_blank;
