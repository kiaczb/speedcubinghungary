"use client";

import { useEffect, useState } from "react";
import {
  FaExclamationTriangle,
  FaListUl,
  FaRegCalendarAlt,
} from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import CompetitionList from "./CompetitionList";
import MonthCalendar from "./MonthCalendar";
import StateMessage from "./StateMessage";

type Status = "loading" | "error" | "success";
type ViewMode = "list" | "calendar";

function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
}

const CardSkeleton = () => (
  <div className="flex animate-pulse flex-col gap-5 rounded-2xl border border-gray-200 bg-white p-5 sm:flex-row sm:gap-6">
    <div className="h-16 w-16 shrink-0 rounded-xl bg-gray-100" />
    <div className="flex flex-1 flex-col gap-3">
      <div className="h-5 w-2/3 rounded bg-gray-100" />
      <div className="h-4 w-1/3 rounded bg-gray-100" />
      <div className="h-4 w-1/2 rounded bg-gray-100" />
      <div className="mt-1 h-9 w-48 rounded-lg bg-gray-100" />
    </div>
  </div>
);

const ViewToggle = ({
  view,
  onChange,
}: {
  view: ViewMode;
  onChange: (view: ViewMode) => void;
}) => {
  const baseClass =
    "inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors";
  const activeClass = "bg-gray-900 text-white";
  const inactiveClass = "text-gray-600 hover:text-gray-900";

  return (
    <div className="inline-flex rounded-lg border border-gray-200 bg-white p-1">
      <button
        type="button"
        onClick={() => onChange("list")}
        aria-pressed={view === "list"}
        className={`${baseClass} ${view === "list" ? activeClass : inactiveClass}`}
      >
        <FaListUl className="h-3.5 w-3.5" />
        Lista
      </button>
      <button
        type="button"
        onClick={() => onChange("calendar")}
        aria-pressed={view === "calendar"}
        className={`${baseClass} ${view === "calendar" ? activeClass : inactiveClass}`}
      >
        <FaRegCalendarAlt className="h-3.5 w-3.5" />
        Naptár
      </button>
    </div>
  );
};

const CalendarView = () => {
  const [status, setStatus] = useState<Status>("loading");
  const [competitions, setCompetitions] = useState<UpcomingCompetition[]>([]);
  const [view, setView] = useState<ViewMode>("list");

  useEffect(() => {
    const controller = new AbortController();

    const fetchCompetitions = async () => {
      setStatus("loading");
      try {
        const res = await fetch(`${getApiBaseUrl()}/competitions/upcoming`, {
          signal: controller.signal,
        });
        if (!res.ok) throw new Error(`Request failed: ${res.status}`);

        const data: UpcomingCompetition[] = await res.json();
        // Defensive: ensure nearest upcoming date is first.
        data.sort((a, b) => a.startDate.localeCompare(b.startDate));

        setCompetitions(data);
        setStatus("success");
      } catch (error) {
        if (controller.signal.aborted) return;
        console.error("Versenyek lekérése sikertelen:", error);
        setStatus("error");
      }
    };

    fetchCompetitions();
    return () => controller.abort();
  }, []);

  let content: React.ReactNode;
  if (status === "loading") {
    content = (
      <div className="flex flex-col gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  } else if (status === "error") {
    content = (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="Nem sikerült betölteni a versenyeket"
        description="Hiba történt az adatok lekérése közben. Kérjük, próbáld újra később."
      />
    );
  } else if (view === "list") {
    content = <CompetitionList competitions={competitions} />;
  } else {
    content = <MonthCalendar competitions={competitions} />;
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex justify-end">
        <ViewToggle view={view} onChange={setView} />
      </div>
      {content}
    </div>
  );
};

export default CalendarView;
