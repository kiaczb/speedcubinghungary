"use client";

import { FaAlignLeft, FaHashtag, FaRegClock, FaUser } from "react-icons/fa";
import Field, { controlClass } from "@/app/components/ui/Field";
import type { CompetitorOptions } from "../lib/useCompetitorOptions";
import { ExtraQuestion } from "../types";
import CompetitorInput from "./CompetitorInput";
import TimeInput from "./TimeInput";
import { parseTimeInput } from "@/app/lib/time";

type Props = {
  question: ExtraQuestion;
  value: string;
  onChange: (value: string) => void;
  /** A verseny teljes nevezői listája — csak a "person" típusú kérdés használja. */
  competitors: CompetitorOptions;
};

const TYPE_ICON = {
  person: FaUser,
  time: FaRegClock,
  number: FaHashtag,
  text: FaAlignLeft,
} as const;

const ExtraQuestionField = ({
  question,
  value,
  onChange,
  competitors,
}: Props) => {
  const Icon = TYPE_ICON[question.type];
  const inputId = `extra-question-${question.id}`;
  const inputClass = `${controlClass} pl-9`;

  return (
    <Field
      label={question.prompt}
      htmlFor={inputId}
      hint={
        question.type === "person" && competitors.status === "error" ? (
          "Nem sikerült betölteni a nevezőket — a nevet kézzel is beírhatod."
        ) : question.type === "time" && value !== "" && parseTimeInput(value) === null ? (
          <span role="alert" className="text-brand-red">
            Érvénytelen idő — perc mellett a másodperc legfeljebb 59 lehet.
          </span>
        ) : undefined
      }
    >
      <div className="relative">
        <Icon className="pointer-events-none absolute left-3.5 top-1/2 z-10 h-3.5 w-3.5 -translate-y-1/2 text-subtle-foreground" />
        {question.type === "person" ? (
          <CompetitorInput
            {...competitors}
            id={inputId}
            value={value}
            onChange={onChange}
            className={inputClass}
          />
        ) : question.type === "time" ? (
          <TimeInput
            id={inputId}
            value={value}
            onChange={onChange}
            className={inputClass}
          />
        ) : question.type === "number" ? (
          <input
            id={inputId}
            type="number"
            inputMode="numeric"
            value={value}
            onChange={(event) => onChange(event.target.value)}
            placeholder="pl. 3"
            className={inputClass}
          />
        ) : (
          <input
            id={inputId}
            type="text"
            value={value}
            onChange={(event) => onChange(event.target.value)}
            placeholder="Válaszod"
            className={inputClass}
          />
        )}
      </div>
    </Field>
  );
};

export default ExtraQuestionField;
