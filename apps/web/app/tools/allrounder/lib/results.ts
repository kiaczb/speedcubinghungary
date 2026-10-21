import type { EventResult } from "../types";

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

export function medalForRank(rank: number): Medal | null {
  if (rank === 1) return "gold";
  if (rank === 2) return "silver";
  if (rank === 3) return "bronze";
  return null;
}

const MEDAL_STYLES: Record<Medal, { accent: string; text: string }> = {
  gold: {
    accent: "bg-[#d4af37]",
    text: "text-[#9a7c1f]",
  },
  silver: {
    accent: "bg-[#9aa0a8]",
    text: "text-[#71767d]",
  },
  bronze: {
    accent: "bg-[#b0752f]",
    text: "text-[#8a5a22]",
  },
};

export function medalAccentClass(medal: Medal): string {
  return MEDAL_STYLES[medal].accent;
}

export function medalTextClass(medal: Medal): string {
  return MEDAL_STYLES[medal].text;
}
