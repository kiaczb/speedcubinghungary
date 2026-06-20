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
    setSelectedDate(new Date(today.getFullYear(), today.getMonth(), today.getDate()));
  };

  const monthLabel = viewMonth.toLocaleDateString("hu-HU", {
    year: "numeric",
    month: "long",
  });

  const selectedCompetitions = selectedDate
    ? getCompetitionsForDay(eventMap, selectedDate)
    : [];

  return (
    <div className="flex flex-col gap-6">
      {/* Month header + navigation */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold capitalize text-gray-900">
          {monthLabel}
        </h2>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={goToToday}
            className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
          >
            Ma
          </button>
          <button
            type="button"
            onClick={goToPreviousMonth}
            aria-label="Előző hónap"
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 text-gray-600 transition-colors hover:bg-gray-50"
          >
            <FaChevronLeft className="h-3.5 w-3.5" />
          </button>
          <button
            type="button"
            onClick={goToNextMonth}
            aria-label="Következő hónap"
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 text-gray-600 transition-colors hover:bg-gray-50"
          >
            <FaChevronRight className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      {/* Calendar grid */}
      <div className="overflow-hidden rounded-2xl border border-gray-200 bg-gray-200">
        {/* Weekday header */}
        <div className="grid grid-cols-7 gap-px bg-gray-200">
          {WEEKDAY_LABELS.map((label) => (
            <div
              key={label}
              className="bg-gray-50 px-1 py-2 text-center text-xs font-medium uppercase tracking-wide text-gray-400"
            >
              {label}
            </div>
          ))}
        </div>

        {/* Day cells */}
        <div className="grid grid-cols-7 gap-px bg-gray-200">
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
                onClick={() => setSelectedDate(new Date(date))}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    setSelectedDate(new Date(date));
                  }
                }}
                className={`relative min-h-14 cursor-pointer p-1.5 text-left transition-colors sm:min-h-28 ${
                  inMonth ? "bg-white hover:bg-gray-50" : "bg-gray-50"
                } ${isSelected ? "ring-2 ring-inset ring-gray-900" : ""}`}
              >
                {/* Day number + mobile indicator */}
                <div className="flex items-center justify-between">
                  <span
                    className={`inline-flex h-6 w-6 items-center justify-center rounded-full text-xs sm:h-7 sm:w-7 sm:text-sm ${
                      isToday
                        ? "bg-gray-900 font-semibold text-white"
                        : inMonth
                          ? "text-gray-900"
                          : "text-gray-300"
                    }`}
                  >
                    {date.getDate()}
                  </span>
                  {hasComps && (
                    <span className="sm:hidden">
                      <span className="flex h-4 min-w-4 items-center justify-center rounded-full bg-gray-900 px-1 text-[10px] font-semibold text-white">
                        {dayComps.length}
                      </span>
                    </span>
                  )}
                </div>

                {/* Desktop competition chips */}
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
                        className="truncate rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-700 transition-colors hover:bg-gray-200"
                      >
                        {competition.name}
                      </a>
                    ))}
                    {dayComps.length > MAX_CHIPS && (
                      <span className="px-1 text-xs text-gray-400">
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

      {/* Selected day agenda */}
      <div>
        {selectedDate ? (
          <>
            <h3 className="mb-3 text-sm font-semibold capitalize text-gray-900">
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
              <p className="rounded-2xl border border-dashed border-gray-300 bg-white px-4 py-6 text-center text-sm text-gray-500">
                Ezen a napon nincs verseny.
              </p>
            )}
          </>
        ) : (
          <p className="rounded-2xl border border-dashed border-gray-300 bg-white px-4 py-6 text-center text-sm text-gray-500">
            Válassz egy napot a versenyek megtekintéséhez.
          </p>
        )}
      </div>
    </div>
  );
};

export default MonthCalendar;
