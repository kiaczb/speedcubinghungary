import { WeeklyReport } from "../types";

/**
 * A riport poszt `content`-je JSON (lásd apps/scripts/weekly-report). Biztonságos
 * parse: hibás/régi tartalomnál `null`, hogy a hívó fallbackelni tudjon és ne
 * dőljön el a renderelés.
 */
export function parseWeeklyReport(content: string): WeeklyReport | null {
  try {
    const data = JSON.parse(content);
    if (
      data &&
      Array.isArray(data.competitions) &&
      Array.isArray(data.competitorMilestones) &&
      Array.isArray(data.delegateMilestones)
    ) {
      return data as WeeklyReport;
    }
    return null;
  } catch {
    return null;
  }
}

/** Rövid szöveges összefoglaló a riportról a listákhoz/kártyákhoz (a JSON-t nem
 *  mutatjuk nyersen). */
export function reportSummary(content: string): string {
  const report = parseWeeklyReport(content);
  if (!report) return "Heti riport";

  const competitions = report.competitions.length;
  const records = report.competitions.reduce(
    (sum, comp) =>
      sum + comp.competitors.reduce((s, c) => s + c.results.length, 0),
    0,
  );

  if (competitions === 0) {
    return "Ezen a héten nem volt magyar érdekeltségű versenyeredmény.";
  }
  return `${competitions} verseny, ${records} rekord ezen a héten.`;
}

/** Kétbetűs ország-ISO → zászló emoji (pl. "HU" → 🇭🇺). */
export function flagEmoji(iso: string): string {
  if (!iso || iso.length !== 2) return iso;
  return [...iso.toUpperCase()]
    .map((c) => String.fromCodePoint(127397 + c.charCodeAt(0)))
    .join("");
}

/** "YYYY-MM-DD" (esetleg tartomány) → magyaros megjelenítés. Stringből dolgozunk,
 *  hogy ne csússzon időzóna-eltolás (a Date az ISO dátumot UTC-ként értené). */
export function formatPeriod(from: string, to: string): string {
  const fmt = (d: string) => d.replaceAll("-", ". ") + ".";
  return from === to ? fmt(from) : `${fmt(from)} – ${fmt(to)}`;
}
