/**
 * Dátumformázás — egy helyen, magyar nyelven, MAGYAR IDŐZÓNÁBAN.
 *
 * ⚠️ Az időzóna rögzítése nem kozmetika. A `toLocaleDateString` a FUTTATÓ
 * környezet zónáját használja: a Server Componentek a konténerben (UTC), a
 * client komponensek a látogató gépén futnak, így ugyanaz a dátum két különböző
 * napnak látszhat, és éjfél közelében hidratálási eltérést okoz.
 *
 * Ha új megjelenítési alak kell, ide vegyél fel egy függvényt; ne írj a
 * komponensbe újabb `toLocaleDateString`-et.
 */

const LOCALE = "hu-HU";
const TIME_ZONE = "Europe/Budapest";

export type DateInput = Date | string | number;

function formatter(options: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
  return new Intl.DateTimeFormat(LOCALE, { timeZone: TIME_ZONE, ...options });
}

const LONG_DATE = formatter({ year: "numeric", month: "long", day: "numeric" });
const LONG_DATE_WITH_WEEKDAY = formatter({
  year: "numeric",
  month: "long",
  day: "numeric",
  weekday: "long",
});
const SHORT_DATE = formatter({
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});
const MONTH_LABEL = formatter({ year: "numeric", month: "long" });
const MONTH_ABBREVIATION = formatter({ month: "short" });
const DAY_OF_MONTH = formatter({ day: "numeric" });
const DATE_TIME = formatter({
  year: "numeric",
  month: "long",
  day: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});
const TIMESTAMP = formatter({ dateStyle: "short", timeStyle: "short" });

/**
 * Érvénytelen dátumon az Intl RangeError-t DOB (a `toLocaleDateString` csak
 * "Invalid Date"-et adott vissza). Egy hibás API-mező ne döntsön ki egy
 * oldalt — üres szöveget adunk helyette.
 */
function format(intl: Intl.DateTimeFormat, value: DateInput): string {
  const date = value instanceof Date ? value : new Date(value);
  return Number.isNaN(date.getTime()) ? "" : intl.format(date);
}

/** `2026. szeptember 24.` — a részletező felületek alapértelmezése. */
export function formatLongDate(value: DateInput): string {
  return format(LONG_DATE, value);
}

/** `2026. szeptember 24., csütörtök` */
export function formatLongDateWithWeekday(value: DateInput): string {
  return format(LONG_DATE_WITH_WEEKDAY, value);
}

/** `2026. 09. 24.` — listákhoz, ahol a dátum másodlagos. */
export function formatShortDate(value: DateInput): string {
  return format(SHORT_DATE, value);
}

/** `2026. szeptember` — a naptár fejléce. */
export function formatMonthLabel(value: DateInput): string {
  return format(MONTH_LABEL, value);
}

/** `SZEPT` — a versenykártya dátumjelvényének felső sora. */
export function formatMonthBadge(value: DateInput): string {
  return format(MONTH_ABBREVIATION, value).replace(".", "").toUpperCase();
}

/** `24` — a versenykártya dátumjelvényének alsó sora. */
export function formatDayOfMonth(value: DateInput): string {
  return format(DAY_OF_MONTH, value);
}

/** `2026. szeptember 24. 18:00` — határidőkhöz, ahol az óra is számít. */
export function formatDateTime(value: DateInput): string {
  return format(DATE_TIME, value);
}

/** `2026. 09. 24. 18:00` — gépi időbélyeghez (pl. „frissítve"). */
export function formatTimestamp(value: DateInput): string {
  return format(TIMESTAMP, value);
}

/** `2026. szeptember 24. – 2026. szeptember 25.`, egynapos versenynél egy dátum. */
export function formatDateRange(start: DateInput, end: DateInput): string {
  const from = formatLongDate(start);
  const to = formatLongDate(end);

  return from === to ? from : `${from} – ${to}`;
}
