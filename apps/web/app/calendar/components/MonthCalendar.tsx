"use client";

import { useMemo, useState } from "react";
import { FaChevronLeft, FaChevronRight } from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import {
  buildEventMap,
  getCompetitionsForDay,
  getInitialMonth,
  getMonthMatrix,
  isSameDay,
  isSameMonth,
  toDateKey,
  WEEKDAY_LABELS,
} from "../lib/calendarEvents";
import CompetitionCard from "./CompetitionCard";

type Props = {
  competitions: UpcomingCompetition[];
};

const MAX_CHIPS = 2;

const AgendaNote = ({ children }: { children: React.ReactNode }) => (
  <p className="rounded-card border border-dashed border-line-strong bg-surface px-4 py-6 text-center text-sm text-muted-foreground">
    {children}
  </p>
);

const MonthCalendar = ({ competitions }: Props) => {
  const eventMap = useMemo(() => buildEventMap(competitions), [competitions]);
  const today = useMemo(() => new Date(), []);

  const [viewMonth, setViewMonth] = useState<Date>(() =>
    getInitialMonth(competitions, today),
  );
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);

  const matrix = useMemo(
    () => getMonthMatrix(viewMonth.getFullYear(), viewMonth.getMonth()),
    [viewMonth],
  );

  const goToPreviousMonth = () => {
    setSelectedDate(null);
    setViewMonth((m) => new Date(m.getFullYear(), m.getMonth() - 1, 1));
  };
  const goToNextMonth = () => {
    setSelectedDate(null);
    setViewMonth((m) => new Date(m.getFullYear(), m.getMonth() + 1, 1));
  };
  const goToToday = () => {
    setViewMonth(new Date(today.getFullYear(), today.getMonth(), 1));
    setSelectedDate(
      new Date(today.getFullYear(), today.getMonth(), today.getDate()),
    );
  };

  const monthLabel = viewMonth.toLocaleDateString("hu-HU", {
    year: "numeric",
    month: "long",
  });

  const selectedCompetitions = selectedDate
    ? getCompetitionsForDay(eventMap, selectedDate)
    : [];

  const navButtonClass =
    "inline-flex h-9 w-9 items-center justify-center rounded-control border border-line text-muted-foreground transition-colors duration-200 ease-smooth hover:border-line-strong hover:bg-surface-muted hover:text-foreground";

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold capitalize text-foreground">
          {monthLabel}
        </h2>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={goToToday}
            className="rounded-control border border-line px-3 py-1.5 text-sm font-medium text-foreground transition-colors duration-200 ease-smooth hover:border-line-strong hover:bg-surface-muted"
          >
            Ma
          </button>
          <button
            type="button"
            onClick={goToPreviousMonth}
            aria-label="Előző hónap"
            className={navButtonClass}
          >
            <FaChevronLeft className="h-3 w-3" />
          </button>
          <button
            type="button"
            onClick={goToNextMonth}
            aria-label="Következő hónap"
            className={navButtonClass}
          >
            <FaChevronRight className="h-3 w-3" />
          </button>
        </div>
      </div>

      <div className="overflow-hidden rounded-card border border-line bg-line">
        <div className="grid grid-cols-7 gap-px bg-line">
          {WEEKDAY_LABELS.map((label) => (
            <div
              key={label}
              className="bg-surface-muted px-1 py-2.5 text-center font-mono text-[10px] font-medium uppercase tracking-[0.14em] text-subtle-foreground"
            >
              {label}
            </div>
          ))}
        </div>

        <div className="grid grid-cols-7 gap-px bg-line">
          {matrix.flat().map((date) => {
            const dayComps = getCompetitionsForDay(eventMap, date);
            const inMonth = isSameMonth(date, viewMonth);
            const isToday = isSameDay(date, today);
            const isSelected = selectedDate
              ? isSameDay(date, selectedDate)
              : false;
            const hasComps = dayComps.length > 0;

            return (
              <div
                key={toDateKey(date)}
                role="button"
                tabIndex={0}
                aria-pressed={isSelected}
                aria-label={date.toLocaleDateString("hu-HU", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
                onClick={() => setSelectedDate(new Date(date))}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    setSelectedDate(new Date(date));
                  }
                }}
                className={`relative min-h-14 cursor-pointer p-1.5 text-left transition-colors duration-200 ease-smooth focus-visible:z-10 sm:min-h-28 ${
                  inMonth ? "bg-surface hover:bg-surface-muted" : "bg-surface-muted/60"
                } ${isSelected ? "ring-2 ring-inset ring-foreground" : ""}`}
              >
                <div className="flex items-center justify-between">
                  <span
                    className={`tabular inline-flex h-6 w-6 items-center justify-center rounded-full font-mono text-xs sm:h-7 sm:w-7 sm:text-sm ${
                      isToday
                        ? "bg-brand-red font-semibold text-white"
                        : inMonth
                          ? "text-foreground"
                          : "text-subtle-foreground/50"
                    }`}
                  >
                    {date.getDate()}
                  </span>
                  {hasComps && (
                    <span className="flex h-4 min-w-4 items-center justify-center rounded-full bg-foreground px-1 font-mono text-[10px] font-semibold text-white sm:hidden">
                      {dayComps.length}
                    </span>
                  )}
                </div>

                {hasComps && (
                  <div className="mt-1 hidden flex-col gap-1 sm:flex">
                    {dayComps.slice(0, MAX_CHIPS).map((competition) => (
                      <a
                        key={competition.id}
                        href={competition.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={(e) => e.stopPropagation()}
                        title={competition.name}
                        className="truncate rounded border-l-2 border-brand-red bg-surface-muted px-1.5 py-0.5 text-xs text-foreground transition-colors duration-200 ease-smooth hover:bg-line"
                      >
                        {competition.name}
                      </a>
                    ))}
                    {dayComps.length > MAX_CHIPS && (
                      <span className="px-1 text-xs text-subtle-foreground">
                        +{dayComps.length - MAX_CHIPS} további
                      </span>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div>
        {selectedDate ? (
          <>
            <h3 className="mb-3 text-sm font-semibold capitalize text-foreground">
              {selectedDate.toLocaleDateString("hu-HU", {
                year: "numeric",
                month: "long",
                day: "numeric",
                weekday: "long",
              })}
            </h3>
            {selectedCompetitions.length > 0 ? (
              <div className="flex flex-col gap-4">
                {selectedCompetitions.map((competition) => (
                  <CompetitionCard
                    key={competition.id}
                    competition={competition}
                  />
                ))}
              </div>
            ) : (
              <AgendaNote>Ezen a napon nincs verseny.</AgendaNote>
            )}
          </>
        ) : (
          <AgendaNote>
            Válassz egy napot a versenyek megtekintéséhez.
          </AgendaNote>
        )}
      </div>
    </div>
  );
};

export default MonthCalendar;
