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
      "bg-sky-700 text-white ring-sky-600/20 dark:bg-sky-950 dark:text-white dark:ring-sky-400",
  },
  event: {
    label: "Esemény",
    color:
      "bg-emerald-700 text-white ring-emerald-600/20 dark:bg-emerald-950 dark:text-white dark:ring-emerald-400",
  },
  competition: {
    label: "Verseny",
    color:
      "bg-rose-700 text-white ring-rose-600/20 dark:bg-rose-950 dark:text-white dark:ring-rose-400",
  },
  announcement: {
    label: "Közlemény",
    color:
      "bg-amber-700 text-white ring-amber-600/25 dark:bg-amber-950 dark:text-white dark:ring-amber-400",
  },
  report: {
    label: "Heti riport",
    color:
      "bg-violet-700 text-white ring-violet-600/20 dark:bg-violet-950 dark:text-white dark:ring-violet-400",
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
