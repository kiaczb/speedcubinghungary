// A /learn/* API válaszainak alakja (apps/api/internal/learn/model.go).
// A backend camelCase-t ad, mint a `special`/`allrounder` végpontok.

export type LearnEvent = {
  id: string;
  name: string;
  description: string;
  order: number;
  /** Csak listázásnál töltött (a részletes válaszokban is jelen van). */
  caseCount: number;
};

export type LearnCase = {
  id: string;
  eventId: string;
  name: string;
  description: string;
  order: number;
  algorithmCount: number;
};

export type LearnAlgorithm = {
  id: string;
  caseId: string;
  name: string;
  notation: string;
  description: string;
  /** Üres string, ha az algoritmushoz nincs videó. */
  videoUrl: string;
  order: number;
};

export type LearnEventDetail = {
  event: LearnEvent;
  cases: LearnCase[];
};

export type LearnCaseDetail = {
  event: LearnEvent;
  case: LearnCase;
  algorithms: LearnAlgorithm[];
};

export type LearnAlgorithmDetail = {
  event: LearnEvent;
  case: LearnCase;
  algorithm: LearnAlgorithm;
  /** Az eset ÖSSZES algoritmusa (a kiválasztottal együtt) — a lapozáshoz. */
  siblings: LearnAlgorithm[];
};

// A teljes fa az admin felületnek (GET /learn/tree): a gyerekek beágyazva.
export type LearnCaseNode = LearnCase & {
  algorithms: LearnAlgorithm[];
};

export type LearnEventNode = LearnEvent & {
  cases: LearnCaseNode[];
};
