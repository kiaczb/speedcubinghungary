import { UpcomingCompetition } from "../types";

// Converts "2026-08-15" into the compact "20260815" form Google Calendar expects.
function toCompactDate(date: string): string {
  return date.replaceAll("-", "");
}

// Returns a new YYYY-MM-DD string shifted by the given number of days.
function shiftDate(date: string, days: number): string {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d.toISOString().split("T")[0];
}

/**
 * Builds a Google Calendar "create event" URL for an all-day competition.
 *
 * Google treats the end date of an all-day event as exclusive, so we add one
 * day to endDate to make the final competition day part of the event.
 */
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
