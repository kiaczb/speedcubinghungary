"use client";

import { controlClass } from "@/app/components/ui/Field";
import { formatTimeInput } from "@/app/lib/time";

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
  return (
    <input
      id={id}
      type="text"
      inputMode="numeric"
      value={value}
      onChange={(event) => {
        // Elutasított bevitelnél (null) a mező az ELŐZŐ értéken marad.
        const formatted = formatTimeInput(event.target.value);
        if (formatted !== null) onChange(formatted);
      }}
      placeholder={placeholder}
      aria-label={ariaLabel}
      autoComplete="off"
      className={className}
    />
  );
};

export default TimeInput;
