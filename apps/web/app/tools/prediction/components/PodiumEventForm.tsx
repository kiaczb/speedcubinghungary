"use client";

import "@cubing/icons";
import { eventName } from "@/app/lib/events";
import { useCompetitorOptions } from "../lib/useCompetitorOptions";
import CompetitorInput from "./CompetitorInput";

type Props = {
  competitionId: string;
  eventId: string;
  picks: [string, string, string];
  onChange: (place: 0 | 1 | 2, name: string) => void;
};

const PLACES = [
  { place: 1, label: "1. hely" },
  { place: 2, label: "2. hely" },
  { place: 3, label: "3. hely" },
] as const;

const PodiumEventForm = ({
  competitionId,
  eventId,
  picks,
  onChange,
}: Props) => {
  const competitors = useCompetitorOptions(
    `/competitions/${encodeURIComponent(competitionId)}/psych-sheet/${encodeURIComponent(eventId)}`,
  );

  return (
    <div className="rounded-control border border-line bg-surface-muted p-4">
      <div className="flex items-center gap-2">
        <span
          className={`cubing-icon event-${eventId} text-lg leading-none text-foreground`}
          aria-hidden
        />
        <h3 className="font-semibold text-foreground">{eventName(eventId)}</h3>
      </div>

      <div className="mt-3 flex flex-col gap-2.5">
        {PLACES.map(({ place, label }) => (
          <div key={place} className="flex items-center gap-3">
            <span className="tabular flex h-9 w-9 shrink-0 items-center justify-center rounded-control border border-line bg-surface font-mono text-sm font-semibold text-muted-foreground">
              {place}.
            </span>
            <div className="min-w-0 flex-1">
              <CompetitorInput
                {...competitors}
                value={picks[place - 1]}
                onChange={(name) => onChange((place - 1) as 0 | 1 | 2, name)}
                ariaLabel={`${eventName(eventId)} ${label}`}
              />
            </div>
          </div>
        ))}
      </div>

      {competitors.status === "error" && (
        <p className="mt-2.5 text-xs text-brand-red-strong">
          Nem sikerült betölteni a versenyzőket — a nevet kézzel is beírhatod.
        </p>
      )}
    </div>
  );
};

export default PodiumEventForm;
