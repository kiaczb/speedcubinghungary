export type PostsItem = {
  id: string;
  tag: TagTypes;
  title: string;
  content: string;
  /**
   * A poszt képeinek API-relatív útvonalai (pl. "/uploads/<uuid>.webp") a
   * MEGJELENÍTÉS sorrendjében — az első elem a borítókép, a többi a poszt
   * oldalán lévő carousel további képe. Az API kép nélküli poszthoz []-t ad.
   */
  image_paths?: string[];
  created_at: string;
};

export const badgeMap = {
  news: {
    label: "Hír",
    color:
      "bg-sky-50 text-sky-700 ring-sky-600/20 dark:bg-sky-950 dark:text-sky-300 dark:ring-sky-400/25",
  },
  event: {
    label: "Esemény",
    color:
      "bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950 dark:text-emerald-300 dark:ring-emerald-400/25",
  },
  competition: {
    label: "Verseny",
    color:
      "bg-rose-50 text-rose-700 ring-rose-600/20 dark:bg-rose-950 dark:text-rose-300 dark:ring-rose-400/25",
  },
  announcement: {
    label: "Közlemény",
    color:
      "bg-amber-50 text-amber-800 ring-amber-600/25 dark:bg-amber-950 dark:text-amber-300 dark:ring-amber-400/25",
  },
  // A weekly-report job által generált heti riport (post_tag 'report').
  // Kötelező bejegyzés: a badge-et a tag alapján keressük ki, enélkül a
  // riport posztok renderelése elszállna.
  report: {
    label: "Heti riport",
    color:
      "bg-violet-50 text-violet-700 ring-violet-600/20 dark:bg-violet-950 dark:text-violet-300 dark:ring-violet-400/25",
  },
} as const;

export type TagTypes = keyof typeof badgeMap;

// --- Weekly report (tag === "report") -----------------------------------------
// A riport posztok `content`-je JSON ezzel a struktúrával (a weekly-report job
// állítja elő, lásd apps/scripts/weekly-report/reportPayload.py). A frontend a
// WeeklyReportView komponenssel rendereli, nem markdownként.

export type WeeklyReportResult = {
  event: string;
  type: string; // "single" | "average"
  result: string; // már formázva a Python oldalon (pl. "6.42")
  badge: string; // "NR" | "CR" | "WR" | "PR"
};

export type WeeklyReportCompetitor = {
  name: string;
  results: WeeklyReportResult[]; // üres, ha a versenyzőnek nincs rekordja
};

export type WeeklyReportCompetition = {
  name: string;
  countryIso: string;
  from: string; // ISO dátum (YYYY-MM-DD)
  to: string;
  competitors: WeeklyReportCompetitor[];
};

export type WeeklyReportMilestoneCompetitor = {
  name: string;
  count: number;
  estimated: string | null;
  estimatedDate: string | null;
  achieved: boolean;
};

export type WeeklyReportMilestoneDelegate = {
  name: string;
  count: number;
};

export type WeeklyReport = {
  competitions: WeeklyReportCompetition[];
  competitorMilestones: WeeklyReportMilestoneCompetitor[];
  delegateMilestones: WeeklyReportMilestoneDelegate[];
};
