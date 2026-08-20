// A Budapest Special kliensoldali típusai. A mezőnevek az API camelCase
// válaszát követik (apps/api/internal/special/model.go).

export type Competition = {
  competitionWcaId: string;
  liveCompetitionId: number;
  name: string;
  startDate: string;
  endDate: string;
  /** A helyszín országa a WCA Live venues mezőjéből, ha ismert. */
  countryIso2?: string;
};

export type CompetitorAttempt = {
  attemptNumber: number;
  /** Hivatalos WCIF AttemptResult: -1 DNF, -2 DNS, 0 kihagyott, >0 századmásodperc. */
  result: number;
  /** A rögzített TÉNYLEGES idő (emberi megfigyelés). */
  elapsedCs?: number;
  /** Felső korlát, nem mérés. */
  maxElapsedCs?: number;
};

export type CompetitorRound = {
  /** A WCA Live result id-ja — a kísérlet sorszámával együtt az írás kulcsa. */
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

/** Egy kumulatív keret — több körre is szólhat (WCIF cumulativeRoundIds). */
export type CumulativeBudget = {
  roundWcifIds: string[];
  limitCs: number;
  completedCs: number;
  recordedDnfCs: number;
  usedCs: number;
  remainingCs: number;
  /** Amíg > 0, a remainingCs csak felső becslés. */
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
  /** A riport tárgya: a néző által választott verseny. */
  competition: Competition;
  /**
   * A versenyen előforduló körsorszámok, növekvő sorrendben. MINDIG a teljes
   * adathalmazból — a kör-választó ebből épül, ezért a szűrés nem szűkíti.
   */
  availableRounds: number[];
  /** A szűrés tárgya; null, ha minden kör benne van. */
  round: number | null;
  competitors: Competitor[];
};
