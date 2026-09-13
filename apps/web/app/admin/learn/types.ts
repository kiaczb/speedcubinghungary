import type {
  LearnAlgorithm,
  LearnCase,
  LearnEvent,
} from "@/app/learn/types";

// A szerkesztő modal állapota. Diszkriminált unió, mert a három szint mezői és
// a szülő-hivatkozásuk eltérnek — `entity: null` a létrehozás, egyébként
// szerkesztés.
export type LearnModalState =
  | { kind: "event"; entity: LearnEvent | null }
  | { kind: "case"; entity: LearnCase | null; eventId: string }
  | { kind: "algorithm"; entity: LearnAlgorithm | null; caseId: string };

export type LearnFeedback = {
  type: "success" | "error";
  message: string;
};
