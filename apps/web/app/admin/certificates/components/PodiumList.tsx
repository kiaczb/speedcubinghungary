import "@cubing/icons";
import { eventName } from "@/app/lib/events";
import { countryName } from "../../stats/lib/format";
import type { CertificateEvent, Recipient } from "../types";
import { placeLabel, resultTypeLabel } from "../lib/recipients";

const STATUS_LABELS = {
  running: "a döntő még fut",
  pending: "a döntő még nem kezdődött el",
};

function EventIcon({ eventId }: { eventId: string }) {
  return (
    <span
      className={`cubing-icon event-${eventId} text-base leading-none text-muted-foreground`}
      aria-hidden
    />
  );
}

export default function PodiumList({
  events,
  recipients,
  isExcluded,
  onToggle,
  isEventExcluded,
  onToggleEvent,
}: {
  events: CertificateEvent[];
  recipients: Recipient[];
  isExcluded: (recipient: Recipient) => boolean;
  onToggle: (recipient: Recipient) => void;
  isEventExcluded: (eventId: string) => boolean;
  onToggleEvent: (eventId: string) => void;
}) {
  const finished = events.filter((e) => e.status === "finished");
  const unfinished = events.filter((e) => e.status !== "finished");

  return (
    <div className="flex flex-col gap-6">
      {unfinished.length > 0 && (
        <div className="rounded-card border border-dashed border-line-strong bg-surface px-4 py-3">
          <p className="text-sm font-medium text-foreground">
            Ezekhez még nem készülhet oklevél
          </p>
          <ul className="mt-2 flex flex-wrap gap-x-5 gap-y-1.5">
            {unfinished.map((event) => (
              <li key={event.eventId} className="flex items-center gap-2 text-sm text-muted-foreground">
                <EventIcon eventId={event.eventId} />
                {eventName(event.eventId)}
                <span className="text-subtle-foreground">
                  · {STATUS_LABELS[event.status as keyof typeof STATUS_LABELS]}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {finished.map((event) => {
        const podium = recipients.filter((r) => r.eventId === event.eventId);
        const eventIncluded = !isEventExcluded(event.eventId);
        return (
          <section key={event.eventId}>
            <label className="flex cursor-pointer items-center gap-3 pb-2">
              <input
                type="checkbox"
                checked={eventIncluded}
                onChange={() => onToggleEvent(event.eventId)}
                disabled={podium.length === 0}
                className="h-4 w-4 accent-foreground"
              />
              <EventIcon eventId={event.eventId} />
              <span className="font-semibold text-foreground">{eventName(event.eventId)}</span>
            </label>

            {podium.length === 0 ? (
              <p className="rounded-card border border-line bg-surface px-4 py-3 text-sm text-muted-foreground">
                Nincs érvényes eredménnyel dobogós.
              </p>
            ) : (
              <ul
                className={`flex flex-col divide-y divide-line rounded-card border border-line bg-surface ${
                  eventIncluded ? "" : "opacity-50"
                }`}
              >
                {podium.map((recipient, index) => {
                  const included = eventIncluded && !isExcluded(recipient);
                  return (
                    <li key={`${recipient.wcaId || recipient.name}-${index}`}>
                      <label
                        className={`flex items-center gap-4 px-4 py-2.5 transition-colors duration-200 ease-smooth ${
                          eventIncluded ? "cursor-pointer hover:bg-surface-muted" : ""
                        } ${eventIncluded && !included ? "opacity-50" : ""}`}
                      >
                        <input
                          type="checkbox"
                          checked={included}
                          disabled={!eventIncluded}
                          onChange={() => onToggle(recipient)}
                          className="h-4 w-4 shrink-0 accent-foreground"
                        />
                        <span className="tabular w-6 shrink-0 font-mono text-sm font-semibold text-foreground">
                          {recipient.place ? placeLabel(recipient.place) : ""}
                        </span>
                        <span className="min-w-0 flex-1">
                          <span className="block truncate font-medium text-foreground">{recipient.name}</span>
                          <span className="tabular block font-mono text-xs text-muted-foreground">
                            {recipient.wcaId || "újonc"}
                            <span className="font-sans"> · {countryName(recipient.countryIso2)}</span>
                          </span>
                        </span>
                        <span className="shrink-0 text-right">
                          <span className="tabular block font-mono text-sm text-foreground">{recipient.result}</span>
                          <span className="block text-xs text-subtle-foreground">
                            {resultTypeLabel(recipient.resultType)}
                          </span>
                        </span>
                      </label>
                    </li>
                  );
                })}
              </ul>
            )}
          </section>
        );
      })}
    </div>
  );
}
