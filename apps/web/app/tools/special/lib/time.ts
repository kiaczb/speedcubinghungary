// Századmásodperc ↔ olvasható idő. A WCA a századmásodpercet használja
// mértékegységként (WCIF AttemptResult), ezért mindenhol egész cs-ben számolunk
// — lebegőpontos másodperc soha nem kerül a rendszerbe.

/** 3742 → "37.42", 12345 → "2:03.45" */
export function formatCentiseconds(centiseconds: number): string {
  const negative = centiseconds < 0;
  const value = Math.abs(centiseconds);

  const minutes = Math.floor(value / 6000);
  const seconds = Math.floor((value % 6000) / 100);
  const hundredths = value % 100;

  const body =
    minutes > 0
      ? `${minutes}:${String(seconds).padStart(2, "0")}.${String(hundredths).padStart(2, "0")}`
      : `${seconds}.${String(hundredths).padStart(2, "0")}`;

  return negative ? `-${body}` : body;
}

/**
 * Beviteli maszk: a beírt számjegyeket jobbról tölti fel, ahogy a versenyeken
 * használt időbeviteli mezők. "3742" → "37.42", "12345" → "1:23.45".
 */
export function formatTimeInput(value: string): string {
  const digits = value.replace(/\D/g, "").replace(/^0+(?=\d)/, "");
  if (!digits) return "";

  if (digits.length <= 2) {
    return `0.${digits.padStart(2, "0")}`;
  }

  if (digits.length <= 4) {
    return `${parseInt(digits.slice(0, -2), 10)}.${digits.slice(-2)}`;
  }

  const minutes = parseInt(digits.slice(0, -4), 10);
  return `${minutes}:${digits.slice(-4, -2)}.${digits.slice(-2)}`;
}

/**
 * "37.42" → 3742, "1:23.45" → 8345. Érvénytelen vagy üres bemenetre null.
 */
export function parseTimeToCentiseconds(formatted: string): number | null {
  const trimmed = formatted.trim().replace(",", ".");
  if (!trimmed) return null;

  const match = trimmed.match(/^(?:(\d+):)?(\d{1,2})\.(\d{1,2})$/);
  if (!match) return null;

  const [, minutes, seconds, fraction] = match;
  const total =
    (minutes ? parseInt(minutes, 10) * 6000 : 0) +
    parseInt(seconds, 10) * 100 +
    parseInt(fraction.padEnd(2, "0"), 10);

  return total > 0 ? total : null;
}

/**
 * Egy hivatalos WCIF AttemptResult megjelenítése.
 * A speciális értékek: -1 DNF, -2 DNS, 0 még nem történt meg / kihagyott.
 */
export function formatAttemptResult(result: number): string {
  if (result === -1) return "DNF";
  if (result === -2) return "DNS";
  if (result === 0) return "—";
  return formatCentiseconds(result);
}
