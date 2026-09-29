-- Védőháló a hitelesítés legfontosabb invariánsára: a users.wca_id a
-- felhasználó azonosítója (UNIQUE kulcs, a jogosultságok hordozója), ezért
-- soha nem lehet üres.
--
-- Miért kell ez DB szinten is: a WCA OAuth callback korábban nem ellenőrizte a
-- WCA válaszainak állapotkódját, így egy hibás (4xx) válasz üres wca_id-vel is
-- végigfutott, és EGYETLEN, mindenki által osztott "szellemfelhasználót" hozott
-- létre üres wca_id-vel — akire a rendszer érvényes munkamenet-sütit állított
-- ki. A hibát az alkalmazás oldalán javítottuk (internal/auth/service.go), ez a
-- megszorítás pedig azt garantálja, hogy a hiba osztálya a tárolási rétegen se
-- legyen többé lehetséges.

-- 1) Ha egy korábbi futásból maradt ilyen sor, ELŐBB tegyük ártalmatlanná.
--    Nem törlünk: a sorra hivatkozhatnak idegen kulcsok (prediction_contests.
--    created_by, special_competitions.added_by). Átnevezve viszont sosem
--    illeszkedik többé egy valódi WCA bejelentkezésre, mert a WCA azonosítók
--    formátuma ettől eltér.
UPDATE users
SET    wca_id = 'INVALID-' || id,
       updated_at = NOW()
WHERE  btrim(wca_id) = '';

-- 2) A megszorítás: sem üres, sem csak whitespace nem lehet (ugyanaz a szabály,
--    amit a Go oldal is alkalmaz a WCA /me válaszára).
ALTER TABLE users
    ADD CONSTRAINT users_wca_id_not_blank CHECK (btrim(wca_id) <> '');
