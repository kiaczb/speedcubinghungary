-- A Special követés egyszerűsítése: a hivatalos eredmények NEM kerülnek
-- adatbázisba, csak az emberi megfigyelés.
--
-- Miért: a special_results tábla a WCA Live adatának szó szerinti másolata volt
-- (snapshotResults → INSERT → SELECT → BuildCompetitors, változtatás nélkül).
-- Ehelyett a riport a WCA Live-ból élőben áll össze, és amíg valaki nézi az
-- oldalt, a roundUpdated subscription tartja frissen. Így nincs mit szinkronban
-- tartani, és megszűnik a "figyelés indítása előtti körök hiányoznak" hibaosztály.
DROP TABLE IF EXISTS special_results;

-- A megfigyelés-tábla csak azt tartja meg, amit a WCA Live NEM tud megmondani:
-- a DNF solve tényleges hosszát. Minden más oszlop származtatott volt, és
-- kérésenként újraszámolható a friss WCA Live adatból:
--
--   event_id / round_number / round_wcif_id / registrant_id / competitor_*
--     → a friss kör-pillanatképben benne van
--   time_limit_cs / cumulative_round_ids / max_elapsed_cs / completed_cs
--     → a limitből és a kísérletekből számolódik (attemptBound)
--   detected_at / entered_at  → felderítés kellett hozzá; már nincs felderítés
--   resolved_at → ha az upstream kísérlet már nem DNF, a sor egyszerűen nem
--     kapcsolódik be a riportba; nem kell külön jelölni
--   live_round_id → a live_result_id globálisan egyedi, nem kell mellé kör
ALTER TABLE special_dnf_attempts
    DROP COLUMN IF EXISTS live_round_id,
    DROP COLUMN IF EXISTS event_id,
    DROP COLUMN IF EXISTS round_number,
    DROP COLUMN IF EXISTS round_wcif_id,
    DROP COLUMN IF EXISTS registrant_id,
    DROP COLUMN IF EXISTS competitor_wca_id,
    DROP COLUMN IF EXISTS competitor_name,
    DROP COLUMN IF EXISTS time_limit_cs,
    DROP COLUMN IF EXISTS cumulative_round_ids,
    DROP COLUMN IF EXISTS max_elapsed_cs,
    DROP COLUMN IF EXISTS completed_cs,
    DROP COLUMN IF EXISTS detected_at,
    DROP COLUMN IF EXISTS entered_at,
    DROP COLUMN IF EXISTS resolved_at;

-- Sor csak akkor keletkezik, ha valaki rögzített időt — üres megfigyelés
-- nincs értelme, a DNF-ek listája a friss WCA Live adatból jön.
DELETE FROM special_dnf_attempts WHERE elapsed_cs IS NULL;

ALTER TABLE special_dnf_attempts
    ALTER COLUMN elapsed_cs SET NOT NULL;

-- A riport egy verseny összes rögzített idejét egyben kéri le.
CREATE INDEX IF NOT EXISTS special_dnf_attempts_competition_idx2
    ON special_dnf_attempts (competition_wca_id);

DROP INDEX IF EXISTS special_dnf_attempts_competition_idx;
DROP INDEX IF EXISTS special_dnf_attempts_round_registrant_idx;
