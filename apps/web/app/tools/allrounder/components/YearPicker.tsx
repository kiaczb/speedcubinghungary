"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { controlClass } from "@/app/components/ui/Field";

type YearPickerProps = {
  selectedYear: number;
  availableYears: number[];
};

/**
 * Évadválasztó. A kiválasztott évad az URL-ben él (`?year=`), így az oldal
 * megosztható és újratöltésre is megmarad — a `page.tsx` onnan olvassa.
 *
 * A `value` helyi állapot, hogy a legördülő azonnal a választott évet mutassa,
 * miközben a szerver oldal még töltődik. A `page.tsx` az évadra kulcsolva
 * rendereli a komponenst, ezért egy vissza-navigáció friss állapottal indul.
 */
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
