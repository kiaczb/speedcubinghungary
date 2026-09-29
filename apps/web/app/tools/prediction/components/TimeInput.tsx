"use client";

import { controlClass } from "@/app/components/ui/Field";
import { formatTimeInput, parseTimeInput } from "@/app/lib/time";

type Props = {
  value: string;
  onChange: (formatted: string) => void;
  placeholder?: string;
  ariaLabel?: string;
  id?: string;
  className?: string;
};


const TimeInput = ({
  value,
  onChange,
  placeholder = "pl. 8.42",
  ariaLabel,
  id,
  className = controlClass,
}: Props) => {
  // A formázó engedi a köztes `1:75.00`-féle állapotot (lásd lib/time.ts), ezért
  // az érvényességet külön jelezzük — a szöveges magyarázat a mező alatti
  // hintben van (ExtraQuestionField), itt a keret és az aria-invalid.
  const invalid = value !== "" && parseTimeInput(value) === null;

  return (
    <input
      id={id}
      type="text"
      inputMode="numeric"
      value={value}
      aria-invalid={invalid}
      onChange={(event) => {
        // Elutasított bevitelnél (null) a mező az ELŐZŐ értéken marad.
        const formatted = formatTimeInput(event.target.value);
        if (formatted !== null) onChange(formatted);
      }}
      placeholder={placeholder}
      aria-label={ariaLabel}
      autoComplete="off"
      className={invalid ? `${className} border-brand-red focus:border-brand-red` : className}
    />
  );
};

export default TimeInput;
