export type PostsItem = {
  id: string;
  tag: TagTypes;
  title: string;
  content: string;
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
  report: {
    label: "Heti riport",
    color:
      "bg-violet-50 text-violet-700 ring-violet-600/20 dark:bg-violet-950 dark:text-violet-300 dark:ring-violet-400/25",
  },
} as const;

export type TagTypes = keyof typeof badgeMap;

export type WeeklyReportResult = {
  event: string;
  type: string;
  result: string;
  badge: string;
};

export type WeeklyReportCompetitor = {
  name: string;
  results: WeeklyReportResult[];
};

export type WeeklyReportCompetition = {
  name: string;
  countryIso: string;
  from: string;
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
