import { FaHashtag, FaRegClock, FaUser } from "react-icons/fa";
import Field, { controlClass } from "@/app/components/ui/Field";
import { ExtraQuestion } from "../types";

type Props = {
  question: ExtraQuestion;
};

const TYPE_ICON = {
  person: FaUser,
  time: FaRegClock,
  number: FaHashtag,
} as const;

const TYPE_PLACEHOLDER = {
  person: "Versenyző neve",
  time: "pl. 8.42",
  number: "pl. 3",
} as const;

// Csak a megjelenés — a mező egyelőre nincs állapothoz/submithoz kötve.
const ExtraQuestionField = ({ question }: Props) => {
  const Icon = TYPE_ICON[question.type];
  const inputId = `extra-question-${question.id}`;

  return (
    <Field label={question.prompt} htmlFor={inputId}>
      <div className="relative">
        <Icon className="pointer-events-none absolute left-3.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-subtle-foreground" />
        <input
          id={inputId}
          type="text"
          placeholder={TYPE_PLACEHOLDER[question.type]}
          className={`${controlClass} pl-9`}
        />
      </div>
    </Field>
  );
};

export default ExtraQuestionField;
