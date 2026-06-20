import { UpcomingCompetition } from "../types";

// Weekday headers, Monday-first (Hungarian convention).
export const WEEKDAY_LABELS = ["H", "K", "Sze", "Cs", "P", "Szo", "V"];

// Parses a "YYYY-MM-DD" string as a local date, avoiding the UTC parsing of
// `new Date("YYYY-MM-DD")` which can shift the day across timezones.
export function parseLocalDate(dateStr: string): Date {
  const [year, month, day] = dateStr.split("-").map(Number);
  return new Date(year, month - 1, day);
}

// Formats a Date as a "YYYY-MM-DD" key in local time.
export function toDateKey(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

export function isSameMonth(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth();
}

/**
 * Builds a map from calendar day (YYYY-MM-DD) to the competitions occurring on
 * it. A multi-day competition is registered on every day in its
 * [startDate, endDate] range, so spanning events show up on all relevant dates.
 */
export function buildEventMap(
  competitions: UpcomingCompetition[],
): Map<string, UpcomingCompetition[]> {
  const map = new Map<string, UpcomingCompetition[]>();

  for (const competition of competitions) {
    const end = parseLocalDate(competition.endDate);
    const cursor = parseLocalDate(competition.startDate);

    while (cursor <= end) {
      const key = toDateKey(cursor);
      const existing = map.get(key);
      if (existing) {
        existing.push(competition);
      } else {
        map.set(key, [competition]);
      }
      cursor.setDate(cursor.getDate() + 1);
    }
  }

  return map;
}

export function getCompetitionsForDay(
  map: Map<string, UpcomingCompetition[]>,
  date: Date,
): UpcomingCompetition[] {
  return map.get(toDateKey(date)) ?? [];
}

/**
 * Builds the matrix of weeks (Monday-first) covering the given month, including
 * the leading/trailing days from adjacent months needed to fill whole weeks.
 */
export function getMonthMatrix(year: number, month: number): Date[][] {
  const firstOfMonth = new Date(year, month, 1);
  // Convert JS Sunday-first (0=Sun) to Monday-first (0=Mon).
  const startOffset = (firstOfMonth.getDay() + 6) % 7;
  const gridStart = new Date(year, month, 1 - startOffset);

  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const weeks = Math.ceil((startOffset + daysInMonth) / 7);

  const matrix: Date[][] = [];
  const cursor = new Date(gridStart);
  for (let w = 0; w < weeks; w++) {
    const week: Date[] = [];
    for (let d = 0; d < 7; d++) {
      week.push(new Date(cursor));
      cursor.setDate(cursor.getDate() + 1);
    }
    matrix.push(week);
  }
  return matrix;
}

/**
 * Returns the first day of the month that should be shown by default: the month
 * of the earliest competition, or the current month when there are none.
 */
export function getInitialMonth(
  competitions: UpcomingCompetition[],
  today: Date,
): Date {
  if (competitions.length === 0) {
    return new Date(today.getFullYear(), today.getMonth(), 1);
  }
  const earliest = competitions.reduce(
    (min, c) => (c.startDate < min ? c.startDate : min),
    competitions[0].startDate,
  );
  const earliestDate = parseLocalDate(earliest);
  return new Date(earliestDate.getFullYear(), earliestDate.getMonth(), 1);
}
