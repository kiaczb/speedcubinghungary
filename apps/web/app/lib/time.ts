const MAX_DIGITS = 6;

const MAX_CENTISECONDS = 60 * 60 * 100;

export function formatTimeInput(value: string): string | null {
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
