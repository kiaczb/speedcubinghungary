import { UpcomingCompetition } from "../types";

function toCompactDate(date: string): string {
  return date.replaceAll("-", "");
}

function shiftDate(date: string, days: number): string {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d.toISOString().split("T")[0];
}

export function buildGoogleCalendarUrl(competition: UpcomingCompetition): string {
  const start = toCompactDate(competition.startDate);
  const end = toCompactDate(shiftDate(competition.endDate, 1));

  const location = [competition.city, competition.country]
    .filter(Boolean)
    .join(", ");

  const params = new URLSearchParams({
    action: "TEMPLATE",
    text: competition.name,
    dates: `${start}/${end}`,
    details: `Hivatalos versenyoldal: ${competition.url}`,
    location,
  });

  return `https://calendar.google.com/calendar/render?${params.toString()}`;
}
