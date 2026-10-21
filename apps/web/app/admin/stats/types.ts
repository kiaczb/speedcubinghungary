export type Phase = "before" | "during" | "after";

export type CompetitionInfo = {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  city: string;
  countryIso2: string;
  url: string;
  eventIds: string[];
  competitorLimit: number | null;
  resultsPostedAt?: string;
};

export type Person = {
  name: string;
  wcaId?: string;
  countryIso2?: string;
};

export type Count = {
  key: string;
  count: number;
};

export type PersonValue = {
  person: Person;
  value: number;
};

export type RankedResult = {
  person: Person;
  eventId: string;
  type: "single" | "average";
  display: string;
  record?: string;
  worldRanking: number;
  nationalRanking: number;
};

export type Anniversary = {
  years: number;
  startYear: number;
  people: Person[];
};

export type MedalHolder = {
  person: Person;
  gold: number;
  silver: number;
  bronze: number;
};

export type RecordHolder = {
  person: Person;
  world: number;
  continental: number;
  national: number;
};

export type FieldStats = {
  competitors: number;
  limit: number | null;
  newcomers: number;
  foreigners: number;
  countries: Count[];
  genders: Count[];
  events: Count[];
  averageEvents: number;
  allEvents: Person[];
  topRanked: RankedResult[];
  recordHolders: RankedResult[];
  veterans: PersonValue[];
  anniversaries: Anniversary[];
  milestones: PersonValue[];
  mostCompetitions: PersonValue[];
  mostMedals: MedalHolder[];
  mostRecords: RecordHolder[];
  careerSolves: number;
  summariesError?: string;
};

export type HistoryStatus = "unavailable" | "running" | "ready";

export type Comeback = {
  person: Person;
  lastCompetition: { id: string; name: string; date: string };
  gapDays: number;
};

export type HistoryStats = {
  status: HistoryStatus;
  done: number;
  total: number;
  failed: number;
  comebacks: Comeback[];
  firstInCountry: Person[];
  busiestYear: PersonValue[];
};

export type AgeStatus = "ready" | "no-grant" | "forbidden" | "unavailable";

export type AgedPerson = {
  person: Person;
  years: number;
  days: number;
};

export type AgeStats = {
  status: AgeStatus;
  known: number;
  youngest: AgedPerson[];
  oldest: AgedPerson[];
  average: number;
  median: number;
};

export type RecordEntry = {
  person: Person;
  eventId: string;
  type: "single" | "average";
  record: string;
  display: string;
};

export type Highlight = {
  person: Person;
  display: string;
};

export type EventBest = {
  eventId: string;
  single?: Highlight;
  average?: Highlight;
};

export type LiveStats = {
  roundsTotal: number;
  roundsOpen: number;
  roundsFinished: number;
  resultsEntered: number;
  attempts: number;
  dnfs: number;
  records: RecordEntry[];
  prs: number;
  prLeaders: PersonValue[];
  eventBests: EventBest[];
  fetchedAt: string;
};

export type EventWinner = {
  eventId: string;
  person: Person;
  best: string;
  average?: string;
};

export type CountryMedals = {
  countryIso2: string;
  gold: number;
  silver: number;
  bronze: number;
};

export type ResultStats = {
  winners: EventWinner[];
  records: RecordEntry[];
  podiums: MedalHolder[];
  countryPodiums: CountryMedals[];
  solves: number;
  dnfs: number;
  solvingCentiseconds: number;
};

export type StatsReport = {
  competition: CompetitionInfo;
  phase: Phase;
  generatedAt: string;
  field?: FieldStats;
  fieldError?: string;
  history: HistoryStats;
  ages: AgeStats;
  live?: LiveStats;
  liveError?: string;
  results?: ResultStats;
  resultsError?: string;
};
