import { WeeklyReport } from "../types";

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

export function flagEmoji(iso: string): string {
  if (!iso || iso.length !== 2) return iso;
  return [...iso.toUpperCase()]
    .map((c) => String.fromCodePoint(127397 + c.charCodeAt(0)))
    .join("");
}

export function formatPeriod(from: string, to: string): string {
  const fmt = (d: string) => d.replaceAll("-", ". ") + ".";
  return from === to ? fmt(from) : `${fmt(from)} – ${fmt(to)}`;
}
