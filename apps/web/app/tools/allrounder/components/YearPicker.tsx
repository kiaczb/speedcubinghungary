"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { controlClass } from "@/app/components/ui/Field";

type YearPickerProps = {
  selectedYear: number;
  availableYears: number[];
};

const YearPicker = ({ selectedYear, availableYears }: YearPickerProps) => {
  const router = useRouter();
  const [value, setValue] = useState(String(selectedYear));
  const [pending, startTransition] = useTransition();

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const year = event.target.value;
    setValue(year);

    startTransition(() => {
      router.push(`/tools/allrounder?year=${year}`, { scroll: false });
    });
  };

  return (
    <div className="flex items-center gap-2.5">
      <label
        htmlFor="allrounder-year"
        className="text-sm font-medium text-foreground"
      >
        Évad
      </label>
      <select
        id="allrounder-year"
        value={value}
        onChange={handleChange}
        disabled={pending}
        className={`${controlClass} tabular w-auto py-2 font-mono disabled:opacity-60`}
      >
        {availableYears.map((year) => (
          <option key={year} value={year}>
            {year}
          </option>
        ))}
      </select>
    </div>
  );
};

export default YearPicker;
