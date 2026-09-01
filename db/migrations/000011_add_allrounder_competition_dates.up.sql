-- Hungarian All-Rounder: a most FUTÓ állomás élő követéséhez kell tudnunk,
-- mikor van egy állomás.
--
-- Miért: a sorozat állomásainak eredménye a verseny napján még nincs feltöltve
-- a WCA-hoz (a /results ilyenkor 404-et vagy üres tömböt ad), viszont a WCA
-- Live-on percről percre látszik. Az élő pontverseny ezt a két forrást fésüli
-- össze. Ahhoz viszont, hogy ne kérdezgessük a WCA Live-ot MINDEN feltöltetlen
-- állomásra (a sorozat végén több hónapnyi jövőbeli állomás is van, sőt olyan
-- azonosító is, amit a WCA még nem is ismer), tudnunk kell a dátumot: élő
-- követés csak a verseny napján indul.
--
-- Mindhárom oszlop NULLABLE, mert egy még ki sem hirdetett állomásról a WCA API
-- sem tud semmit — a series.go-ban ilyenkor is ott van már a tervezett
-- azonosító.
ALTER TABLE allrounder_syncs
    ADD COLUMN start_date DATE,
    ADD COLUMN end_date   DATE,
    -- A WCA Live numerikus versenyazonosítója. Nincs rá GraphQL lekérdezés, a
    -- /link/competitions/{wcaId} 302-jéből jön — érdemes megjegyezni.
    ADD COLUMN live_competition_id INTEGER;
