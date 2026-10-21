export type PredictionStatus = "open" | "closed" | "evaluated";

export type PredictionContest = {
  id: string;
  competitionId: string;
  competitionName: string;
  competitionCity: string;
  competitionCountry: string;
  competitionStartDate: string;
  predictionDeadline: string;
  status: PredictionStatus;
  participantCount: number;
  podiumEventIds: string[];
  extraQuestions: ExtraQuestion[];
};

export type ExtraQuestionType = "person" | "time" | "number" | "text";

export const extraQuestionTypeLabels: Record<ExtraQuestionType, string> = {
  person: "Versenyző",
  time: "Idő",
  number: "Szám",
  text: "Szabad szöveg",
};

export type ExtraQuestion = {
  id: string;
  prompt: string;
  type: ExtraQuestionType;
};

export type CompetitorOption = {
  name: string;
  wcaId: string;
  position?: number;
};

export const statusBadgeMap: Record<
  PredictionStatus,
  { label: string; color: string }
> = {
  open: {
    label: "Tippelhető",
    color:
      "bg-emerald-50 text-emerald-700 ring-emerald-600/20",
  },
  closed: {
    label: "Lezárva",
    color:
      "bg-amber-50 text-amber-800 ring-amber-600/25",
  },
  evaluated: {
    label: "Kiértékelve",
    color:
      "bg-sky-50 text-sky-700 ring-sky-600/20",
  },
};
