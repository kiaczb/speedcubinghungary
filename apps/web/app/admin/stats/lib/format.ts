const COUNTRY_NAMES = new Intl.DisplayNames(["hu"], { type: "region" });
const NUMBER = new Intl.NumberFormat("hu-HU");

const GENDERS: Record<string, string> = {
  m: "Férfi",
  f: "Nő",
  o: "Egyéb",
};

export function countryName(iso2?: string): string {
  if (!iso2) return "";
  try {
    return COUNTRY_NAMES.of(iso2) ?? iso2;
  } catch {
    return iso2;
  }
}

export function genderName(key: string): string {
  return GENDERS[key] ?? "Nem ismert";
}

export function formatNumber(value: number): string {
  return NUMBER.format(value);
}

export function formatPercent(part: number, whole: number): string {
  if (whole <= 0) return "0%";
  return `${Math.round((part / whole) * 100)}%`;
}

export function formatGap(days: number): string {
  const years = Math.floor(days / 365.25);
  const months = Math.floor((days - years * 365.25) / 30.44);
  if (years === 0) return `${months} hónap`;
  return months > 0 ? `${years} év ${months} hónap` : `${years} év`;
}

export function formatSolvingTime(centiseconds: number): string {
  const minutes = Math.round(centiseconds / 6000);
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  if (hours === 0) return `${rest} perc`;
  return rest > 0 ? `${hours} óra ${rest} perc` : `${hours} óra`;
}

export function resultType(type: "single" | "average"): string {
  return type === "average" ? "átlag" : "single";
}

const BACK_VOWEL_DIGITS = new Set([3, 6, 8]);
const BACK_VOWEL_TENS = new Set([2, 3, 6, 8]);

export function inYear(year: number): string {
  const lastDigit = year % 10;
  if (lastDigit !== 0) {
    return `${year}-${BACK_VOWEL_DIGITS.has(lastDigit) ? "ban" : "ben"}`;
  }
  const tens = Math.floor(year / 10) % 10;
  if (tens !== 0) {
    return `${year}-${BACK_VOWEL_TENS.has(tens) ? "ban" : "ben"}`;
  }
  return `${year}-ben`;
}
