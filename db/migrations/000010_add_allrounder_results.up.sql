-- Hungarian All-Rounder: a WCA hivatalos eredményeinek gyorsítótára.
--
-- Miért tároljuk egyáltalán: a pontverseny egy egész évad (10-12 verseny)
-- ÖSSZES eredményéből számolódik. Eddig ezt a böngésző kérte le
-- versenyenként külön a WCA API-tól minden oldalbetöltéskor (~100 KB/verseny),
-- és ott is számolt. A WCA eredménye viszont a feltöltés után VÉGLEGES —
-- ugyanazt a választ kértük újra és újra. Innentől egyszer kérjük le, itt
-- tároljuk, és a pontszámítás a backendben fut.
--
-- Ez NEM a versenysorozat konfigurációja: melyik verseny tartozik melyik
-- évadhoz, mennyi pontot ér egy helyezés és hány eredmény számít, az továbbra
-- is kódban van (apps/api/internal/allrounder/series.go), ahogy korábban is
-- (utils/competitionIds.ts). Itt kizárólag a WCA-tól letöltött nyers eredmény
-- és a letöltés állapota él.

-- Versenyenkénti szinkron-állapot. Azért külön tábla, mert akkor is kell sor,
-- ha a versenynek MÉG NINCS eredménye (a WCA API ilyenkor 404-et ad): e nélkül
-- minden kérésnél újra megpróbálnánk letölteni a jövőbeli versenyeket is.
CREATE TABLE allrounder_syncs (
    competition_wca_id TEXT PRIMARY KEY,           -- pl. 'HungarianAllRounder12026'
    fetched_at         TIMESTAMPTZ NOT NULL,       -- a legutóbbi PRÓBÁLKOZÁS ideje
    result_count       INTEGER NOT NULL DEFAULT 0,
    -- Ha a legutóbbi letöltés hibára futott, itt marad az üzenet, és a korábbi
    -- eredmények érintetlenül maradnak — hibás WCA-válasz miatt sose ürüljön ki
    -- a tábla. Hibás állapotban rövidebb ideig érvényes a gyorsítótár.
    last_error         TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Egy verseny egy eredménysora, a WCA API /competitions/{id}/results válaszából.
CREATE TABLE allrounder_results (
    competition_wca_id TEXT NOT NULL
        REFERENCES allrounder_syncs(competition_wca_id) ON DELETE CASCADE,
    -- ⚠️ A WCA API válaszTÖMBJÉBEN elfoglalt hely. A pontszámítás VISSZAFELÉ
    -- olvassa a versenyt (így a versenyszám utolsó köre — a döntő — jön előbb),
    -- tehát a sorrend a számítás bemenete, nem dekoráció. Emiatt ez a
    -- kulcs második fele: a sorrendet bitre pontosan visszaadjuk.
    ordinal            INTEGER NOT NULL,

    wca_id             TEXT NOT NULL,              -- '2017TSVE02'
    name               TEXT NOT NULL,
    event_id           TEXT NOT NULL,              -- '333'
    round_type_id      TEXT NOT NULL,              -- '1' | 'f' | 'c' | 'd' | …
    pos                INTEGER NOT NULL,           -- helyezés a körben (a DNF-esek holtversenyben az utolsó helyen)
    best               INTEGER NOT NULL,           -- századmásodperc; negatív = DNF/DNS

    PRIMARY KEY (competition_wca_id, ordinal)
);

-- A tipikus lekérdezés egy évad összes versenyének eredményét kéri, sorrendben.
CREATE INDEX allrounder_results_competition_idx
    ON allrounder_results (competition_wca_id, ordinal);
