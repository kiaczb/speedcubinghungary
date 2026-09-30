import "@cubing/icons";
import { eventName } from "@/app/lib/events";
import type { EventResult } from "../types";

type EventChipsProps = {
  results: EventResult[];
};

const EventChips = ({ results }: EventChipsProps) => {
  if (results.length === 0) {
    return (
      <span className="block text-center text-subtle-foreground" aria-hidden>
        –
      </span>
    );
  }

  return (
    <div className="mx-auto grid w-max grid-cols-2 gap-1">
      {results.map((result) => (
        <span
          key={`${result.competitionWcaId}-${result.eventId}`}
          title={`${eventName(result.eventId)} — ${result.ranking}. hely, ${result.points} pont`}
          className="flex items-center gap-1 rounded-control border border-line bg-surface-muted px-1.5 py-1"
        >
          <span
            className={`cubing-icon event-${result.eventId} text-[13px] leading-none text-foreground`}
            aria-hidden
          />
          <span className="tabular font-mono text-xs font-medium leading-none text-foreground">
            {result.ranking}
          </span>
          <span className="sr-only">
            {eventName(result.eventId)}: {result.ranking}. hely
          </span>
        </span>
      ))}
    </div>
  );
};

export default EventChips;
