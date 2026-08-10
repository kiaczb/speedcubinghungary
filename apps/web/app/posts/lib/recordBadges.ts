/**
 * A rekord-jelölések (WR/CR/NR/PR) színei.
 *
 * Ez a térkép szó szerint kétszer szerepelt: a markdown-renderelőben
 * (`PostContent`) és a heti riport nézetben (`WeeklyReportView`). Mivel
 * ugyanazt a fogalmat jelölik, egy helyen élnek — így nem tud elcsúszni, hogy
 * egy NR a poszt szövegében más színű, mint a riport táblázatában.
 *
 * WR = világrekord, CR = kontinensrekord, NR = országos rekord, PR = egyéni rekord.
 */
export const RECORD_BADGE_COLORS: Record<string, string> = {
  WR: "bg-amber-300 text-amber-950",
  CR: "bg-sky-600 text-white",
  NR: "bg-brand-green text-white",
  PR: "bg-surface-muted text-muted-foreground ring-1 ring-inset ring-line",
};

/** Ismeretlen jelölés esetére semleges tónus. */
export const RECORD_BADGE_FALLBACK = "bg-surface-muted text-foreground";
