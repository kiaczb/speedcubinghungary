import { FaExternalLinkAlt } from "react-icons/fa";
import { UpcomingCompetition } from "@/app/calendar/types";
import {
  RegistrationState,
  registrationState,
} from "@/app/calendar/lib/registration";
import { serverJson } from "@/app/lib/api";
import {
  formatDayMonth,
  formatDayMonthTime,
  formatDayOfMonth,
  formatMonthBadge,
} from "@/app/lib/date";
import { Tile } from "./Tile";

const SHOWN = 3;

const DOT_CLASSES: Record<RegistrationState["kind"], string> = {
  upcoming: "bg-line-strong",
  open: "bg-brand-green",
  full: "bg-brand-red",
  closed: "bg-line-strong",
  unknown: "",
};

function registrationText(state: RegistrationState): string | null {
  switch (state.kind) {
    case "upcoming":
      return `Nevezés nyílik: ${formatDayMonthTime(state.opensAt)}`;
    case "open":
      return [
        "Nevezés nyitva",
        state.spotsLeft !== null ? `még ${state.spotsLeft} hely` : null,
        `zárul ${formatDayMonth(state.closesAt)}`,
      ]
        .filter(Boolean)
        .join(" · ");
    case "full":
      return `Betelt · várólista · zárul ${formatDayMonth(state.closesAt)}`;
    case "closed":
      return "Nevezés lezárult";
    case "unknown":
      return null;
  }
}

function RegistrationLine({ state }: { state: RegistrationState }) {
  const text = registrationText(state);
  if (!text) return null;

  const emphasized = state.kind === "open" || state.kind === "full";

  return (
    <span
      className={`mt-1.5 flex items-center gap-2 text-xs ${
        emphasized ? "font-medium text-foreground" : "text-subtle-foreground"
      }`}
    >
      <span
        className={`h-1.5 w-1.5 shrink-0 rounded-full ${DOT_CLASSES[state.kind]}`}
        aria-hidden="true"
      />
      {text}
    </span>
  );
}

export default async function UpcomingCompetitionsTile({
  className = "",
}: {
  className?: string;
}) {
  const competitions = await serverJson<UpcomingCompetition[]>(
    "/competitions/upcoming",
    [],
  );
  const now = new Date();
  const shown = competitions.slice(0, SHOWN);

  return (
    <Tile
      href="/calendar"
      title="Közelgő versenyek"
      cta="Versenynaptár"
      linkWholeCard={false}
      className={className}
    >
      {shown.length === 0 ? (
        <p className="text-sm leading-relaxed text-muted-foreground">
          Most nincs meghirdetett hazai verseny. Amint kiírunk egyet, itt
          jelenik meg.
        </p>
      ) : (
        <ul className="-mx-2 flex flex-col gap-1">
          {shown.map((competition) => (
            <li key={competition.id}>
              <a
                href={competition.url}
                target="_blank"
                rel="noopener noreferrer"
                className="group/row flex items-start gap-3 rounded-control px-2 py-2 transition-colors duration-200 ease-smooth hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
              >
                <span className="flex w-9 shrink-0 flex-col items-center pt-0.5">
                  <span className="font-mono text-[10px] font-medium uppercase text-subtle-foreground">
                    {formatMonthBadge(competition.startDate)}
                  </span>
                  <span className="tabular font-mono text-lg font-bold leading-tight text-foreground">
                    {formatDayOfMonth(competition.startDate)}
                  </span>
                </span>
                <span className="min-w-0 flex-1">
                  <span className="flex items-center gap-1.5">
                    <span className="truncate text-sm font-semibold text-foreground">
                      {competition.name}
                    </span>
                    <FaExternalLinkAlt
                      className="h-2.5 w-2.5 shrink-0 text-subtle-foreground opacity-0 transition-opacity duration-200 ease-smooth group-hover/row:opacity-100"
                      aria-hidden="true"
                    />
                  </span>
                  <span className="block text-xs text-muted-foreground">
                    {competition.city}
                  </span>
                  <RegistrationLine
                    state={registrationState(competition, now)}
                  />
                </span>
              </a>
            </li>
          ))}
        </ul>
      )}
    </Tile>
  );
}
