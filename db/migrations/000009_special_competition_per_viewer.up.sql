-- A Special versenyét ezentúl a NÉZŐ választja ki, nem az admin.
--
-- Miért: a figyelés eddig szerveroldali állapot volt (pontosan egy aktív sor),
-- így egyszerre egyetlen versenyt lehetett nézni, és ahhoz admin beavatkozás
-- kellett. Az eredmények viszont a WCA Live-ból élőben jönnek, a kliens pedig
-- kérésenként megadja, melyik versenyt nézi (?competition=…) — az "aktív
-- verseny" fogalma ezzel értelmét vesztette.
--
-- A tábla megmarad, de más szerepben: a special_dnf_attempts idegen kulcsának
-- szülője, vagyis azoknak a versenyeknek a nyilvántartása, amelyekhez EMBER
-- rögzített DNF-időt. Sor továbbra is csak ilyenkor keletkezik, publikus
-- kérésre soha.
ALTER TABLE special_competitions
    DROP COLUMN IF EXISTS active;
