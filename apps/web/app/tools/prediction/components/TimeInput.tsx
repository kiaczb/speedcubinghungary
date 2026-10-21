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
  const invalid = value !== "" && parseTimeInput(value) === null;

  return (
    <input
      id={id}
      type="text"
      inputMode="numeric"
      value={value}
      aria-invalid={invalid}
      onChange={(event) => {
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
