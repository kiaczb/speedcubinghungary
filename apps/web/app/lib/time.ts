/**
 * WCA-stílusú időbevitel: a felhasználó csak számjegyeket ír, és azokat
 * jobbról tagoljuk század/másodperc/perc szerint (`14289` → `1:42.89`,
 * `842` → `8.42`, `389` → `3.89`). Semmilyen más karakter nem kerül be — ez
 * zárja ki, hogy a felhasználó félkész `:`/`.` jeleket írjon, amiket aztán
 * újra kellene értelmezni.
 *
 * ⚠️ Ez KÖZÖS, mert a tippverseny és a Budapest Special ugyanezt csinálta,
 * két külön, viselkedésében ELTÉRŐ megvalósítással: a Special változata nem
 * ismert korlátot, így hat számjegy fölött `123:45.67`-féle értéket is
 * elfogadott — olyan időt, ami egy kockaversenyen nem létezik, és amit a
 * mentés csak a szerveren bukott volna el. A megtartott (szigorúbb) változat a
 * tippversenyé.
 */

const MAX_DIGITS = 6;

/** 60:00.00 — a leghosszabb WCA versenyszám (333mbf) felső határa. */
const MAX_CENTISECONDS = 60 * 60 * 100;

/**
 * `null` = a bevitel nem fogadható el (túl hosszú vagy 60 perc fölött). A
 * hívó ilyenkor HAGYJA a mezőt az előző értéken — ne ugorjon el, és ne
 * csonkoljon némán.
 */
export function formatTimeInput(value: string): string | null {
  // A vezető nullákat le KELL vágni: a megjelenített `0.01`-ből a nyers
  // számjegyek `001`-et adnak, és enélkül a következő leütéssel
  // `0014` → `00.14` → … → `001:23.45` halmozódna.
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
