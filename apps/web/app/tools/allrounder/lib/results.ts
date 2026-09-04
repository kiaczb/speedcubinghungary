import type { EventResult } from "../types";

/**
 * Egy versenyző beszámító eredményei állomásonként csoportosítva. A táblázat
 * (asztali) és a lista (mobil) nézet ugyanezt a csoportosítást használja.
 */
export function groupByCompetition(
  results: EventResult[],
): Map<string, EventResult[]> {
  const grouped = new Map<string, EventResult[]>();

  for (const result of results) {
    const existing = grouped.get(result.competitionWcaId);
    if (existing) {
      existing.push(result);
    } else {
      grouped.set(result.competitionWcaId, [result]);
    }
  }

  return grouped;
}

export type Medal = "gold" | "silver" | "bronze";

/**
 * Az 1–3. helyezés érme-színe. Csak a szoros dobogóra (top 3) vonatkozik —
 * a `podiumCount` (a sorozat végén díjazott helyezettek száma) ettől
 * független és lehet annál nagyobb is; azokra a helyekre marad a semleges
 * kiemelés.
 */
export function medalForRank(rank: number): Medal | null {
  if (rank === 1) return "gold";
  if (rank === 2) return "silver";
  if (rank === 3) return "bronze";
  return null;
}

// Az érmék tényleges, ismert színei — nem márkaszín, ezért nincs saját
// design tokenjük és nem invertálódnak sötét módban, csak kicsit
// felfényesednek a jobb kontrasztért sötét háttéren.
const MEDAL_STYLES: Record<Medal, { accent: string; text: string }> = {
  gold: {
    accent: "bg-[#d4af37] dark:bg-[#e8c766]",
    text: "text-[#9a7c1f] dark:text-[#e8c766]",
  },
  silver: {
    accent: "bg-[#9aa0a8] dark:bg-[#c2c8d0]",
    text: "text-[#71767d] dark:text-[#c2c8d0]",
  },
  bronze: {
    accent: "bg-[#b0752f] dark:bg-[#d4915a]",
    text: "text-[#8a5a22] dark:text-[#d4915a]",
  },
};

export function medalAccentClass(medal: Medal): string {
  return MEDAL_STYLES[medal].accent;
}

export function medalTextClass(medal: Medal): string {
  return MEDAL_STYLES[medal].text;
}
