import { FaPlus, FaTrash } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import { controlClass } from "@/app/components/ui/Field";
import {
  ExtraQuestion,
  ExtraQuestionType,
  extraQuestionTypeLabels,
} from "../types";

type Props = {
  questions: ExtraQuestion[];
  onChange: (questions: ExtraQuestion[]) => void;
};

let nextId = 0;
function newQuestionId(): string {
  nextId += 1;
  return `new-${Date.now()}-${nextId}`;
}

const ExtraQuestionsEditor = ({ questions, onChange }: Props) => {
  const addQuestion = () => {
    onChange([
      ...questions,
      { id: newQuestionId(), prompt: "", type: "number" },
    ]);
  };

  const updateQuestion = (id: string, patch: Partial<ExtraQuestion>) => {
    onChange(questions.map((q) => (q.id === id ? { ...q, ...patch } : q)));
  };

  const removeQuestion = (id: string) => {
    onChange(questions.filter((q) => q.id !== id));
  };

  return (
    <div className="flex flex-col gap-3">
      {questions.length > 0 && (
        <ul className="flex flex-col gap-2.5">
          {questions.map((question, index) => (
            <li
              key={question.id}
              className="flex flex-col gap-2 rounded-control border border-line bg-surface-muted p-3 sm:flex-row sm:items-center"
            >
              <input
                type="text"
                value={question.prompt}
                onChange={(e) =>
                  updateQuestion(question.id, { prompt: e.target.value })
                }
                placeholder={`${index + 1}. kérdés szövege`}
                aria-label={`${index + 1}. kérdés szövege`}
                className={`${controlClass} sm:flex-1`}
              />

              <select
                value={question.type}
                onChange={(e) =>
                  updateQuestion(question.id, {
                    type: e.target.value as ExtraQuestionType,
                  })
                }
                aria-label={`${index + 1}. kérdés típusa`}
                className={`${controlClass} sm:w-40`}
              >
                {Object.entries(extraQuestionTypeLabels).map(
                  ([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ),
                )}
              </select>

              <button
                type="button"
                onClick={() => removeQuestion(question.id)}
                aria-label={`${index + 1}. kérdés törlése`}
                className="inline-flex h-9 w-9 shrink-0 items-center justify-center self-end rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-brand-red/10 hover:text-brand-red sm:self-auto"
              >
                <FaTrash className="h-3.5 w-3.5" />
              </button>
            </li>
          ))}
        </ul>
      )}

      <Button
        type="button"
        variant="secondary"
        size="sm"
        onClick={addQuestion}
        icon={<FaPlus className="h-3 w-3" />}
        className="self-start"
      >
        Kérdés hozzáadása
      </Button>
    </div>
  );
};

export default ExtraQuestionsEditor;
