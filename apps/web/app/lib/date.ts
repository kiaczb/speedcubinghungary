const LOCALE = "hu-HU";
const TIME_ZONE = "Europe/Budapest";

export type DateInput = Date | string | number;

function formatter(options: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
  return new Intl.DateTimeFormat(LOCALE, { timeZone: TIME_ZONE, ...options });
}

const LONG_DATE = formatter({ year: "numeric", month: "long", day: "numeric" });
const LONG_DATE_WITH_WEEKDAY = formatter({
  year: "numeric",
  month: "long",
  day: "numeric",
  weekday: "long",
});
const SHORT_DATE = formatter({
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});
const MONTH_LABEL = formatter({ year: "numeric", month: "long" });
const MONTH_ABBREVIATION = formatter({ month: "short" });
const DAY_OF_MONTH = formatter({ day: "numeric" });
const DATE_TIME = formatter({
  year: "numeric",
  month: "long",
  day: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});
const TIMESTAMP = formatter({ dateStyle: "short", timeStyle: "short" });
const DAY_MONTH = formatter({ month: "short", day: "numeric" });
const ISO_DATE = new Intl.DateTimeFormat("en-CA", {
  timeZone: TIME_ZONE,
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});
const DAY_MONTH_TIME = formatter({
  month: "short",
  day: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

function format(intl: Intl.DateTimeFormat, value: DateInput): string {
  const date = value instanceof Date ? value : new Date(value);
  return Number.isNaN(date.getTime()) ? "" : intl.format(date);
}

export function formatLongDate(value: DateInput): string {
  return format(LONG_DATE, value);
}

export function formatLongDateWithWeekday(value: DateInput): string {
  return format(LONG_DATE_WITH_WEEKDAY, value);
}

export function formatShortDate(value: DateInput): string {
  return format(SHORT_DATE, value);
}

export function formatMonthLabel(value: DateInput): string {
  return format(MONTH_LABEL, value);
}

export function formatMonthBadge(value: DateInput): string {
  return format(MONTH_ABBREVIATION, value).replace(".", "").toUpperCase();
}

export function formatDayOfMonth(value: DateInput): string {
  return format(DAY_OF_MONTH, value);
}

export function formatDateTime(value: DateInput): string {
  return format(DATE_TIME, value);
}

export function formatTimestamp(value: DateInput): string {
  return format(TIMESTAMP, value);
}

export function formatDayMonth(value: DateInput): string {
  return format(DAY_MONTH, value);
}

export function formatDayMonthTime(value: DateInput): string {
  return format(DAY_MONTH_TIME, value);
}

export function formatIsoDate(value: DateInput): string {
  return format(ISO_DATE, value);
}

export function formatDateRange(start: DateInput, end: DateInput): string {
  const from = formatLongDate(start);
  const to = formatLongDate(end);

  return from === to ? from : `${from} – ${to}`;
}
