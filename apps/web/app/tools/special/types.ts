export type Competition = {
  competitionWcaId: string;
  liveCompetitionId: number;
  name: string;
  startDate: string;
  endDate: string;
  countryIso2?: string;
};

export type CompetitorAttempt = {
  attemptNumber: number;
  result: number;
  elapsedCs?: number;
  maxElapsedCs?: number;
};

export type CompetitorRound = {
  liveResultId: number;
  roundWcifId: string;
  eventId: string;
  roundNumber: number;
  attempts: CompetitorAttempt[];
  best?: number;
  average?: number;
  timeLimitCs?: number;
  cumulativeRoundIds: string[];
  completedCs: number;
  recordedDnfCs: number;
  missingDnfCount: number;
};

export type CumulativeBudget = {
  roundWcifIds: string[];
  limitCs: number;
  completedCs: number;
  recordedDnfCs: number;
  usedCs: number;
  remainingCs: number;
  missingDnfCount: number;
};

export type Competitor = {
  registrantId: number;
  competitorWcaId?: string;
  name: string;
  rounds: CompetitorRound[];
  budgets: CumulativeBudget[];
  usedCs: number;
  dnfCount: number;
  missingDnfCount: number;
};

export type SpecialReport = {
  competition: Competition;
  availableRounds: number[];
  round: number | null;
  competitors: Competitor[];
};
