-- Az active oszlop visszaállítása. Az eredeti alapértéket (TRUE) adjuk vissza;
-- melyik verseny volt "aktív", azt nem lehet visszaállítani, ezért a
-- visszaállás után a figyelt versenyt újra be kell állítani.
ALTER TABLE special_competitions
    ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;
