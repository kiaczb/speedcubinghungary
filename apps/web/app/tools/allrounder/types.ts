export type SeriesCompetition = {
  competitionWcaId: string;
  label: string;
};

export type EventResult = {
  competitionWcaId: string;
  eventId: string;
  ranking: number;
  points: number;
};

export type PersonStandings = {
  wcaId: string;
  name: string;
  totalPoints: number;
  results: EventResult[];
};

export type LiveLeg = {
  competitionWcaId: string;
  label: string;
  roundCount: number;
};

export type Standings = {
  year: number;
  availableYears: number[];
  pointsTable: number[];
  countingResults: number;
  podiumCount: number;
  competitions: SeriesCompetition[];
  people: PersonStandings[];
  updatedAt: string | null;
  live?: LiveLeg;
};
