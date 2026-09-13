export type PredictionStatus = "open" | "closed" | "evaluated";

export type PredictionContest = {
  id: string;
  competitionId: string; // WCA verseny id (pl. "HungarianOpen2026")
  competitionName: string;
  competitionCity: string;
  competitionCountry: string;
  /** A tippverseny alapjául szolgáló verseny kezdő dátuma (YYYY-MM-DD). */
  competitionStartDate: string;
  /** Eddig lehet tippelni (ISO dátum). */
  predictionDeadline: string;
  status: PredictionStatus;
  /** Hány felhasználó adott már le tippet. */
  participantCount: number;
  /** Azok a versenyszámok, amikhez dobogó-tippet kell leadni. */
  podiumEventIds: string[];
  /** Admin által megadott extra kérdések (pl. "Hány rekord születik?"). */
  extraQuestions: ExtraQuestion[];
};

// --- Extra kérdések a tippversenyhez -------------------------------------
// Első körben három mezőtípus: versenyző, idő, szám. Ezekből épül fel egy
// szabad kérdés (pl. "Ki nyeri a 3x3x3-at?" → person, "Mennyi lesz a 3x3x3
// országos csúcs?" → time, "Hány rekord születik a versenyen?" → number).

export type ExtraQuestionType = "person" | "time" | "number";

export const extraQuestionTypeLabels: Record<ExtraQuestionType, string> = {
  person: "Versenyző",
  time: "Idő",
  number: "Szám",
};

export type ExtraQuestion = {
  id: string;
  prompt: string;
  type: ExtraQuestionType;
};

// --- Dobogó-tipp form állapota (admin létrehozó modal) --------------------

export type NewPredictionContest = {
  competitionId: string;
  predictionDeadline: string;
  podiumEventIds: string[];
  extraQuestions: ExtraQuestion[];
};

export const statusBadgeMap: Record<
  PredictionStatus,
  { label: string; color: string }
> = {
  open: {
    label: "Tippelhető",
    color:
      "bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-950 dark:text-emerald-300 dark:ring-emerald-400/25",
  },
  closed: {
    label: "Lezárva",
    color:
      "bg-amber-50 text-amber-800 ring-amber-600/25 dark:bg-amber-950 dark:text-amber-300 dark:ring-amber-400/25",
  },
  evaluated: {
    label: "Kiértékelve",
    color:
      "bg-sky-50 text-sky-700 ring-sky-600/20 dark:bg-sky-950 dark:text-sky-300 dark:ring-sky-400/25",
  },
};
