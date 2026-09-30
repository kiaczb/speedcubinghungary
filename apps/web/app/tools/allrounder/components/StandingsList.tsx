"use client";

import { useState } from "react";
import { FaChevronDown } from "react-icons/fa";
import "@cubing/icons";
import { eventName } from "@/app/lib/events";
import { groupByCompetition, medalAccentClass, medalForRank, medalTextClass } from "../lib/results";
import type { PersonStandings, SeriesCompetition } from "../types";

type StandingsListProps = {
  competitions: SeriesCompetition[];
  people: PersonStandings[];
  podiumCount: number;
};

const StandingsList = ({
  competitions,
  people,
  podiumCount,
}: StandingsListProps) => {
  const [openIds, setOpenIds] = useState<ReadonlySet<string>>(new Set());

  const toggle = (wcaId: string) => {
    setOpenIds((previous) => {
      const next = new Set(previous);
      if (!next.delete(wcaId)) next.add(wcaId);
      return next;
    });
  };

  return (
    <ul className="divide-y divide-line overflow-hidden rounded-card border border-line bg-surface">
      {people.map((person, index) => {
        const rank = index + 1;
        const isPodium = rank <= podiumCount;
        const medal = medalForRank(rank);
        const isOpen = openIds.has(person.wcaId);
        const byCompetition = groupByCompetition(person.results);
        const stations = competitions.filter((competition) =>
          byCompetition.has(competition.competitionWcaId),
        );

        return (
          <li key={person.wcaId}>
            <button
              type="button"
              onClick={() => toggle(person.wcaId)}
              aria-expanded={isOpen}
              aria-controls={`allrounder-person-${person.wcaId}`}
              className="relative flex w-full items-center gap-3 px-4 py-3 text-left transition-colors duration-200 ease-smooth hover:bg-surface-muted"
            >
              {isPodium && (
                <span
                  className={`absolute inset-y-0 left-0 w-0.5 ${
                    medal ? medalAccentClass(medal) : "bg-brand-green"
                  }`}
                  aria-hidden
                />
              )}

              <span
                className={`tabular w-6 shrink-0 text-right font-mono text-xs ${
                  medal
                    ? `font-semibold ${medalTextClass(medal)}`
                    : isPodium
                      ? "font-semibold text-foreground"
                      : "text-muted-foreground"
                }`}
              >
                {rank}
              </span>
              <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">
                {person.name}
              </span>
              <span className="tabular shrink-0 font-mono text-sm font-semibold text-foreground">
                {person.totalPoints}
              </span>
              <FaChevronDown
                className={`h-3 w-3 shrink-0 text-subtle-foreground transition-transform duration-200 ease-smooth ${
                  isOpen ? "rotate-180" : ""
                }`}
                aria-hidden
              />
            </button>

            {isOpen && (
              <div
                id={`allrounder-person-${person.wcaId}`}
                className="border-t border-line px-4 pb-4 pt-3"
              >
                {stations.length === 0 ? (
                  <p className="text-xs text-muted-foreground">
                    Ebben az évadban még nincs beszámító eredménye.
                  </p>
                ) : (
                  <dl className="space-y-3">
                    {stations.map((competition) => (
                      <div key={competition.competitionWcaId}>
                        <dt className="font-mono text-xs font-medium uppercase tracking-wide text-muted-foreground">
                          {competition.label}
                        </dt>
                        <dd className="mt-1.5 space-y-1">
                          {(
                            byCompetition.get(competition.competitionWcaId) ?? []
                          ).map((result) => (
                            <div
                              key={result.eventId}
                              className="flex items-center gap-2.5 rounded-control border border-line bg-surface-muted px-2.5 py-1.5"
                            >
                              <span
                                className={`cubing-icon event-${result.eventId} shrink-0 text-base leading-none text-foreground`}
                                aria-hidden
                              />
                              <span className="min-w-0 flex-1 truncate text-xs text-foreground">
                                {eventName(result.eventId)}
                              </span>
                              <span className="tabular shrink-0 font-mono text-xs text-muted-foreground">
                                {result.ranking}. hely
                              </span>
                              <span className="tabular shrink-0 font-mono text-xs font-semibold text-foreground">
                                {result.points} pont
                              </span>
                            </div>
                          ))}
                        </dd>
                      </div>
                    ))}
                  </dl>
                )}
              </div>
            )}
          </li>
        );
      })}
    </ul>
  );
};

export default StandingsList;
