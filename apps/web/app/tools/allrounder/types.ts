/** A backend `/allrounder/standings` válaszának típusai (camelCase). */

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
  /** Csak a BESZÁMÍTÓ eredmények (a legjobb `countingResults` darab). */
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
  /** Az 1., 2., … helyezésért járó pont; a hossza a pontozott mezőny mérete. */
  pointsTable: number[];
  countingResults: number;
  podiumCount: number;
  /** A sorozat összes állomása — az eredmény nélküliek is, oszlopként. */
  competitions: SeriesCompetition[];
  people: PersonStandings[];
  /** A legrégebbi szinkron ideje: ennél régebbi adat nincs a táblázatban. */
  updatedAt: string | null;
  live?: LiveLeg;
};
