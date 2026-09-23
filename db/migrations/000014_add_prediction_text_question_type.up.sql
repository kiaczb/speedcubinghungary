-- Szabadszavas extra kérdés típus (pl. "Milyen új trükköt fogunk látni?") —
-- olyan kérdésekhez, amik nem versenyzőt, időt vagy számot várnak válaszul.
-- PG12+ óta az ADD VALUE tranzakcióban is futhat (az új érték csak a
-- következő tranzakcióban használható — ez a migráció nem is használja).
ALTER TYPE prediction_question_type ADD VALUE IF NOT EXISTS 'text';
