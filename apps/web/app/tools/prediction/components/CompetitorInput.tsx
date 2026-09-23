"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { controlClass } from "@/app/components/ui/Field";
import type { CompetitorOptions } from "../lib/useCompetitorOptions";

type Props = CompetitorOptions & {
  value: string;
  onChange: (name: string) => void;
  ariaLabel?: string;
  id?: string;
  placeholder?: string;
  className?: string;
};


const CompetitorInput = ({
  options,
  status,
  value,
  onChange,
  ariaLabel,
  id,
  placeholder = "Versenyző neve",
  className = controlClass,
}: Props) => {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onClickOutside = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onClickOutside);
    return () => document.removeEventListener("mousedown", onClickOutside);
  }, []);

  const suggestions = useMemo(() => {
    const needle = value.trim().toLowerCase();
    return needle
      ? options.filter((option) => option.name.toLowerCase().includes(needle))
      : options;
  }, [options, value]);

  return (
    <div ref={containerRef} className="relative">
      <input
        id={id}
        type="text"
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        placeholder={status === "loading" ? "Versenyzők betöltése…" : placeholder}
        aria-label={ariaLabel}
        autoComplete="off"
        className={className}
      />

      {open && status === "ready" && suggestions.length > 0 && (
        <ul className="absolute z-10 mt-1.5 max-h-64 w-full overflow-y-auto rounded-control border border-line bg-surface py-1 shadow-card">
          {suggestions.map((option, index) => (
            <li key={option.wcaId || `${option.name}-${index}`}>
              <button
                type="button"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  onChange(option.name);
                  setOpen(false);
                }}
                className="flex w-full items-center justify-between gap-3 px-3.5 py-2 text-left text-sm text-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted"
              >
                <span className="truncate">{option.name}</span>
                {option.position ? (
                  <span className="tabular shrink-0 font-mono text-xs text-subtle-foreground">
                    #{option.position}
                  </span>
                ) : null}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default CompetitorInput;
