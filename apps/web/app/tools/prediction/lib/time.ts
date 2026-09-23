// WCA-stílusú időbevitel: a felhasználó csak számjegyeket ír, és azokat
// jobbról tagoljuk század/másodperc/perc szerint (pl. "14289" → "1:42.89",
// "842" → "8.42", "389" → "3.89"). Nem enged be semmilyen más karaktert —
// ez zárja ki, hogy a felhasználó saját maga írjon félkész ":"/"." jeleket,
// amiket aztán újra kellene értelmezni.

const MAX_DIGITS = 6;
const MAX_CENTISECONDS = 60 * 60 * 100;

// `null` = a bevitel nem fogadható el (túl hosszú vagy 60 perc fölött); a
// mező ilyenkor az előző értéken marad, nem ugrik el.
export function formatTimeInput(digits: string): string | null {
  // A vezető nullákat le KELL vágni: a megjelenített "0.01"-ből az
  // extractDigits "001"-et ad vissza, és a következő gépelt számjeggyel
  // különben "0014" → "00.14" → … → "001:23.45" halmozódna.
  const clean = digits.replace(/\D/g, "").replace(/^0+/, "");
  if (clean === "") return "";
  if (clean.length > MAX_DIGITS) return null;

  const centiseconds = clean.slice(-2).padStart(2, "0");
  const rest = clean.slice(0, -2);
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

// A megjelenített formázott stringből visszaállítja a beírt "nyers"
// számjegyeket, hogy a mező onChange-kor újra tudjuk formázni a friss bevitelt.
export function extractDigits(value: string): string {
  return value.replace(/\D/g, "");
}
