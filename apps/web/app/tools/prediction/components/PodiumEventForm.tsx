import "@cubing/icons";
import { controlClass } from "@/app/components/ui/Field";
import { eventName } from "../lib/events";

type Props = {
  eventId: string;
};

const PLACES = [
  { place: 1, label: "1. hely" },
  { place: 2, label: "2. hely" },
  { place: 3, label: "3. hely" },
];

// Csak a megjelenés — a mezők egyelőre nincsenek állapothoz/submithoz kötve.
const PodiumEventForm = ({ eventId }: Props) => {
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
            <input
              type="text"
              placeholder="Versenyző neve"
              aria-label={`${eventName(eventId)} ${label}`}
              className={controlClass}
            />
          </div>
        ))}
      </div>
    </div>
  );
};

export default PodiumEventForm;
