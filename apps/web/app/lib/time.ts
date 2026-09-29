/**
 * WCA-stílusú időbevitel — a tippverseny és a Budapest Special közös
 * implementációja.
 *
 * Két lépés, két felelősség:
 *   - `formatTimeInput` gépelés KÖZBEN formáz: a felhasználó csak számjegyeket
 *     ír, és azokat jobbról tagoljuk (`14289` → `1:42.89`, `842` → `8.42`).
 *     Más karakter nem kerül be, így nincs félkész `:`/`.`, amit újra kellene
 *     értelmezni.
 *   - `parseTimeInput` dönti el, hogy a kész érték ÉRVÉNYES idő-e. Minden
 *     hívó ezt kérdezi meg, mielőtt az értéket felhasználná vagy elküldené.
 */

const MAX_DIGITS = 6;

/** 60:00.00 — a leghosszabb WCA versenyszám (333mbf) felső határa. */
const MAX_CENTISECONDS = 60 * 60 * 100;

/**
 * `null` = a leütés nem fogadható el (hat számjegy fölött vagy 60 perc
 * fölött). A hívó ilyenkor HAGYJA a mezőt az előző értéken — ne ugorjon el,
 * és ne csonkoljon némán.
 *
 * ⚠️ Azt szándékosan NEM utasítja el, ha perc mellett a másodperc 60 vagy
 * több (`1:75.00`). A számjegyek jobbról töltődnek, így például a `16:05.00`
 * beírása közben az ötjegyű köztes állapot `1:60.50` — ha a formázó ezt
 * elutasítaná, a 16–19., 26–29. … perc sosem lenne begépelhető. Az ilyen
 * érték megjelenhet a mezőben, de a `parseTimeInput` érvénytelennek mondja.
 */
export function formatTimeInput(value: string): string | null {
  // A vezető nullákat le KELL vágni: a megjelenített `0.01` nyers számjegyei
  // `001`, és a következő leütéssel `0014` → `00.14` → … → `001:23.45`
  // halmozódna.
  const digits = value.replace(/\D/g, "").replace(/^0+/, "");
  if (digits === "") return "";
  if (digits.length > MAX_DIGITS) return null;

  const centiseconds = digits.slice(-2).padStart(2, "0");
  const rest = digits.slice(0, -2);
  const seconds = rest.slice(-2);
  const minutes = rest.slice(0, -2);

  const total =
    Number(minutes || 0) * 6000 +
    Number(seconds || 0) * 100 +
    Number(centiseconds);
  if (total > MAX_CENTISECONDS) return null;

  if (rest === "") return `0.${centiseconds}`;
  if (minutes === "") return `${seconds}.${centiseconds}`;
  return `${minutes}:${seconds.padStart(2, "0")}.${centiseconds}`;
}

/**
 * A beírt időt századmásodpercre alakítja; `null`, ha nem érvényes idő.
 *
 * Perc mellett a másodperc legfeljebb 59 lehet — az `1:75.00` elgépelés, nem
 * 2:15.00. Perc nélkül a 99 másodpercig terjedő alak (`74.50`) egyértelmű,
 * ezért elfogadjuk.
 */
export function parseTimeInput(value: string): number | null {
  const trimmed = value.trim().replace(",", ".");
  if (!trimmed) return null;

  const match = trimmed.match(/^(?:(\d+):)?(\d{1,2})\.(\d{1,2})$/);
  if (!match) return null;

  const [, minutes, seconds, fraction] = match;
  if (minutes !== undefined && Number(seconds) >= 60) return null;

  const total =
    (minutes ? Number(minutes) * 6000 : 0) +
    Number(seconds) * 100 +
    Number(fraction.padEnd(2, "0"));

  return total > 0 && total <= MAX_CENTISECONDS ? total : null;
}
