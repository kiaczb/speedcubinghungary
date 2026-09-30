export const DNF = -1;
export const DNS = -2;

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

export function formatAttemptResult(result: number): string {
  if (result === DNF) return "DNF";
  if (result === DNS) return "DNS";
  if (result === 0) return "—";
  return formatCentiseconds(result);
}
