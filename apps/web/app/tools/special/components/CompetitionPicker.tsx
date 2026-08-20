"use client";

import { useEffect, useState } from "react";
import {
  FaCircle,
  FaExclamationTriangle,
  FaSearch,
  FaSpinner,
} from "react-icons/fa";
import { controlClass } from "@/app/components/ui/Field";
import StateMessage from "@/app/components/ui/StateMessage";
import { clientApiUrl } from "@/app/lib/api";
import type { Competition } from "../types";

const SEARCH_DEBOUNCE_MS = 300;

type CompetitionPickerProps = {
  onSelect: (competition: Competition) => void;
};

function isOngoing(competition: Competition, today: string): boolean {
  return competition.startDate <= today && today <= competition.endDate;
}

function formatRange(competition: Competition): string {
  return competition.startDate === competition.endDate
    ? competition.startDate
    : `${competition.startDate} – ${competition.endDate}`;
}

const CompetitionPicker = ({ onSelect }: CompetitionPickerProps) => {
  const [query, setQuery] = useState("");
  const [competitions, setCompetitions] = useState<Competition[]>([]);
  const [status, setStatus] = useState<"loading" | "ready" | "error">(
    "loading",
  );

  useEffect(() => {
    let cancelled = false;
    const needle = query.trim();

    const timer = setTimeout(() => {
      const url = needle
        ? `${clientApiUrl()}/special/competitions?query=${encodeURIComponent(needle)}`
        : `${clientApiUrl()}/special/competitions`;

      fetch(url)
        .then((res) => {
          if (!res.ok) throw new Error(`A kérés sikertelen (${res.status})`);
          return res.json();
        })
        .then((data: Competition[]) => {
          if (cancelled) return;
          setCompetitions(data);
          setStatus("ready");
        })
        .catch((error) => {
          console.error("A versenylista lekérése sikertelen:", error);
          if (!cancelled) setStatus("error");
        });
    }, needle ? SEARCH_DEBOUNCE_MS : 0);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [query]);

  const today = new Date().toISOString().slice(0, 10);

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
          placeholder="Verseny keresése név szerint…"
          aria-label="Verseny keresése"
          className={`${controlClass} pl-9`}
          autoComplete="off"
        />
      </div>

      {status === "loading" ? (
        <StateMessage
          icon={<FaSpinner className="h-5 w-5 animate-spin" />}
          title="Betöltés…"
          description="Versenyek lekérése a WCA Live-ról."
        />
      ) : status === "error" ? (
        <StateMessage
          icon={<FaExclamationTriangle className="h-5 w-5" />}
          title="Nem sikerült betölteni"
          description="A versenylista a WCA Live-ról érkezik. Próbáld újra később."
        />
      ) : competitions.length === 0 ? (
        <StateMessage
          icon={<FaSearch className="h-5 w-5" />}
          title="Nincs találat"
          description={
            query.trim()
              ? "Nincs ilyen nevű verseny a WCA Live-on."
              : "Jelenleg nincs magyar verseny a WCA Live-on. Keress rá a verseny nevére."
          }
        />
      ) : (
        <ul className="flex flex-col gap-2">
          {competitions.map((competition) => {
            const ongoing = isOngoing(competition, today);

            return (
              <li key={competition.competitionWcaId}>
                <button
                  type="button"
                  onClick={() => onSelect(competition)}
                  className="flex w-full items-center justify-between gap-4 rounded-card border border-line bg-surface px-4 py-3 text-left transition-colors duration-200 ease-smooth hover:border-line-strong hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-foreground">
                      {competition.name}
                    </p>
                    <p className="tabular truncate font-mono text-xs text-muted-foreground">
                      {formatRange(competition)}
                      {competition.countryIso2
                        ? ` · ${competition.countryIso2}`
                        : ""}
                    </p>
                  </div>

                  {ongoing && (
                    <span className="flex shrink-0 items-center gap-1.5 text-xs font-medium text-brand-green">
                      <FaCircle className="h-2 w-2" aria-hidden />
                      Most fut
                    </span>
                  )}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
};

export default CompetitionPicker;
