import type { Competitor } from "../types";

export type RoundLimit = {
  key: string;
  limitCs: number;
  label: string;
};

export function roundLabel(roundWcifIds: string[]): string {
  const numbers = new Set(roundWcifIds.map((id) => Number(id.split("-r")[1])));
  const sorted = Array.from(numbers).sort((left, right) => left - right);
  return `${sorted.join(", ")}. kör`;
}

export function collectLimits(competitors: Competitor[]): RoundLimit[] {
  const limits = new Map<string, RoundLimit>();

  for (const competitor of competitors) {
    for (const budget of competitor.budgets) {
      if (budget.roundWcifIds.length < 2) continue;

      const key = budget.roundWcifIds.join("|");
      if (limits.has(key)) continue;

      limits.set(key, {
        key,
        limitCs: budget.limitCs,
        label: roundLabel(budget.roundWcifIds),
      });
    }
  }

  return Array.from(limits.values()).sort((left, right) =>
    left.label.localeCompare(right.label, "hu", { numeric: true }),
  );
}
