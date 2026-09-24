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
import StateMessage from "@/app/components/ui/StateMessage";
import { clientJson } from "@/app/lib/api";

type Status = "loading" | "error" | "success";
type ViewMode = "list" | "calendar";

const CardSkeleton = () => (
  <div className="flex animate-pulse flex-col gap-5 rounded-card border border-line bg-surface p-5 sm:flex-row sm:gap-6">
    <div className="h-16 w-16 shrink-0 rounded-control bg-surface-muted" />
    <div className="flex flex-1 flex-col gap-3">
      <div className="h-5 w-2/3 rounded bg-surface-muted" />
      <div className="h-4 w-1/3 rounded bg-surface-muted" />
      <div className="h-4 w-1/2 rounded bg-surface-muted" />
      <div className="mt-1 h-9 w-48 rounded-control bg-surface-muted" />
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
    "inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors duration-200 ease-smooth";
  const activeClass = "bg-foreground text-background shadow-card";
  const inactiveClass =
    "text-muted-foreground hover:bg-surface-muted hover:text-foreground";

  const options: { value: ViewMode; label: string; icon: React.ReactNode }[] = [
    { value: "list", label: "Lista", icon: <FaListUl className="h-3 w-3" /> },
    {
      value: "calendar",
      label: "Naptár",
      icon: <FaRegCalendarAlt className="h-3 w-3" />,
    },
  ];

  return (
    <div className="inline-flex rounded-control border border-line bg-surface p-1">
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          aria-pressed={view === option.value}
          className={`${baseClass} ${view === option.value ? activeClass : inactiveClass}`}
        >
          {option.icon}
          {option.label}
        </button>
      ))}
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
        const data = await clientJson<UpcomingCompetition[]>(
          "/competitions/upcoming",
          { signal: controller.signal },
        );
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
      <div className="flex items-center justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          {status === "success" && (
            <>
              <span className="tabular font-mono font-medium text-foreground">
                {competitions.length}
              </span>{" "}
              közelgő verseny
            </>
          )}
        </p>
        <ViewToggle view={view} onChange={setView} />
      </div>
      {content}
    </div>
  );
};

export default CalendarView;
