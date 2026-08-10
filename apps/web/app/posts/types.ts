export type PostsItem = {
  id: string;
  tag: TagTypes;
  title: string;
  content: string;
  /** A borítókép API-relatív útvonala (pl. "/uploads/<uuid>.webp"); az API ""-t ad kép nélküli poszthoz. */
  image_path?: string;
  created_at: string;
};

// A `color` teljes tónust ad (háttér + szöveg + gyűrű), nem csak háttérszínt.
// A korábbi telített változatban a sárga "Közlemény" fehér szöveggel gyakorlatilag
// olvashatatlan volt; a lágy tónusok mindegyik címkén megfelelő kontrasztot adnak.
export const badgeMap = {
  news: { label: "Hír", color: "bg-sky-50 text-sky-700 ring-sky-600/20" },
  event: {
    label: "Esemény",
    color: "bg-emerald-50 text-emerald-700 ring-emerald-600/20",
  },
  competition: {
    label: "Verseny",
    color: "bg-rose-50 text-rose-700 ring-rose-600/20",
  },
  announcement: {
    label: "Közlemény",
    color: "bg-amber-50 text-amber-800 ring-amber-600/25",
  },
  // A weekly-report job által generált heti riport (post_tag 'report').
  // Kötelező bejegyzés: a badge-et a tag alapján keressük ki, enélkül a
  // riport posztok renderelése elszállna.
  report: {
    label: "Heti riport",
    color: "bg-violet-50 text-violet-700 ring-violet-600/20",
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
