"use client";

import { useEffect, useState } from "react";
import { FaRegCalendarTimes, FaExclamationTriangle } from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import CompetitionCard from "./CompetitionCard";

type Status = "loading" | "error" | "success";

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

const StateMessage = ({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) => (
  <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-white px-6 py-16 text-center">
    <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-50 text-gray-400">
      {icon}
    </div>
    <h3 className="text-base font-semibold text-gray-900">{title}</h3>
    <p className="mt-1 max-w-sm text-sm text-gray-500">{description}</p>
  </div>
);

const CalendarList = () => {
  const [status, setStatus] = useState<Status>("loading");
  const [competitions, setCompetitions] = useState<UpcomingCompetition[]>([]);

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

  if (status === "loading") {
    return (
      <div className="flex flex-col gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (status === "error") {
    return (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="Nem sikerült betölteni a versenyeket"
        description="Hiba történt az adatok lekérése közben. Kérjük, próbáld újra később."
      />
    );
  }

  if (competitions.length === 0) {
    return (
      <StateMessage
        icon={<FaRegCalendarTimes className="h-5 w-5" />}
        title="Nincs közelgő verseny"
        description="Jelenleg nincs meghirdetett magyar verseny. Nézz vissza később!"
      />
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {competitions.map((competition) => (
        <CompetitionCard key={competition.id} competition={competition} />
      ))}
    </div>
  );
};

export default CalendarList;
