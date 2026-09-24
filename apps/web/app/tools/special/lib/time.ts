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

// A gépelés közbeni formázás a KÖZÖS app/lib/time.ts-ben van (a tippverseny is
// azt használja) — itt csak az marad, ami a WCA eredményekre jellemző.

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

export function formatAttemptResult(result: number): string {
  if (result === -1) return "DNF";
  if (result === -2) return "DNS";
  if (result === 0) return "—";
  return formatCentiseconds(result);
}
