"use client";

import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { FaExclamationTriangle, FaSearch, FaSpinner } from "react-icons/fa";
import { controlClass } from "@/app/components/ui/Field";
import StateMessage from "@/app/components/ui/StateMessage";
import { clientJson } from "@/app/lib/api";
import { formatDateRange, formatIsoDate } from "@/app/lib/date";
import type { CompetitionInfo } from "../types";
import { countryName } from "../lib/format";

const SEARCH_DEBOUNCE_MS = 300;

type Loaded = {
  query: string;
  competitions: CompetitionInfo[];
};

function CompetitionList({
  title,
  competitions,
  onSelect,
}: {
  title: string;
  competitions: CompetitionInfo[];
  onSelect: (id: string) => void;
}) {
  if (competitions.length === 0) return null;

  return (
    <div>
      <h2 className="text-sm font-semibold text-muted-foreground">{title}</h2>
      <ul className="mt-2 flex flex-col divide-y divide-line rounded-card border border-line bg-surface">
        {competitions.map((competition) => (
          <li key={competition.id}>
            <button
              type="button"
              onClick={() => onSelect(competition.id)}
              className="flex w-full flex-col gap-0.5 px-4 py-3 text-left transition-colors duration-200 ease-smooth first:rounded-t-card last:rounded-b-card hover:bg-surface-muted sm:flex-row sm:items-baseline sm:justify-between sm:gap-6"
            >
              <span className="font-medium text-foreground">{competition.name}</span>
              <span className="tabular shrink-0 font-mono text-xs text-muted-foreground">
                {formatDateRange(competition.startDate, competition.endDate)}
                {" · "}
                {competition.city}
                {competition.countryIso2 !== "HU" && `, ${countryName(competition.countryIso2)}`}
              </span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default function CompetitionPicker({
  onSelect,
}: {
  onSelect: (id: string) => void;
}) {
  const [query, setQuery] = useState("");
  const [loaded, setLoaded] = useState<Loaded | null>(null);
  const [failedQuery, setFailedQuery] = useState<string | null>(null);
  const needle = query.trim();

  useEffect(() => {
    let cancelled = false;
    const timer = setTimeout(
      () => {
        const path = needle
          ? `/stats/competitions?query=${encodeURIComponent(needle)}`
          : "/stats/competitions";
        clientJson<CompetitionInfo[]>(path)
          .then((competitions) => {
            if (cancelled) return;
            setLoaded({ query: needle, competitions });
            setFailedQuery(null);
          })
          .catch((error) => {
            console.error("A versenylista lekérése sikertelen:", error);
            if (!cancelled) setFailedQuery(needle);
          });
      },
      needle ? SEARCH_DEBOUNCE_MS : 0,
    );

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [needle]);

  const today = formatIsoDate(new Date());
  const current = loaded?.query === needle ? loaded.competitions : null;

  let body: ReactNode;
  if (failedQuery === needle) {
    body = (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="A versenylista nem érhető el"
        description="A WCA most nem válaszolt. Próbáld újra kicsit később."
      />
    );
  } else if (!current) {
    body = (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Versenyek betöltése…"
        description="A WCA versenylistájából."
      />
    );
  } else if (current.length === 0) {
    body = (
      <p className="text-sm text-muted-foreground">Nincs ilyen nevű verseny.</p>
    );
  } else if (needle) {
    body = (
      <CompetitionList title="Találatok" competitions={current} onSelect={onSelect} />
    );
  } else {
    const upcoming = current.filter((c) => c.endDate >= today);
    const past = current.filter((c) => c.endDate < today).reverse();
    body = (
      <div className="flex flex-col gap-8">
        <CompetitionList title="Közelgő és folyamatban lévő" competitions={upcoming} onSelect={onSelect} />
        <CompetitionList title="Az elmúlt egy év" competitions={past} onSelect={onSelect} />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="relative sm:max-w-md">
        <FaSearch
          className="pointer-events-none absolute left-3.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-subtle-foreground"
          aria-hidden
        />
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Bármely WCA-verseny keresése név szerint…"
          aria-label="Verseny keresése"
          className={`${controlClass} pl-9`}
          autoComplete="off"
        />
      </div>
      {body}
    </div>
  );
}
