import { serverJson } from "@/app/lib/api";
import { Standings } from "@/app/tools/allrounder/types";
import { LiveIndicator, Tile } from "./Tile";

const TOP_COUNT = 8;
const COMPACT_COUNT = 5;

type LegState = "done" | "live" | "pending";

function legStates(standings: Standings): LegState[] {
  const scored = new Set(
    standings.people.flatMap((person) =>
      person.results.map((result) => result.competitionWcaId),
    ),
  );

  return standings.competitions.map(({ competitionWcaId }) => {
    if (standings.live?.competitionWcaId === competitionWcaId) return "live";
    return scored.has(competitionWcaId) ? "done" : "pending";
  });
}

const LEG_CLASSES: Record<LegState, string> = {
  done: "bg-brand-green",
  live: "bg-brand-red motion-safe:animate-pulse",
  pending: "bg-line",
};

export default async function AllRounderTile({
  className = "",
}: {
  className?: string;
}) {
  const standings = await serverJson<Standings | null>(
    "/allrounder/standings",
    null,
  );

  const tileProps = {
    href: "/tools/allrounder",
    cta: "Teljes ranglista",
    className,
  };

  if (!standings || standings.people.length === 0) {
    return (
      <Tile {...tileProps} title="Hungarian All-Rounder">
        <p className="text-sm leading-relaxed text-muted-foreground">
          A sorozat állomásain szerzett helyezések pontokat érnek, és az évad
          végén az összesített pontszám dönt.
        </p>
      </Tile>
    );
  }

  const legs = legStates(standings);
  const doneCount = legs.filter((leg) => leg !== "pending").length;
  const top = standings.people.slice(0, TOP_COUNT);

  return (
    <Tile
      {...tileProps}
      title={`All-Rounder ${standings.year}`}
      meta={
        standings.live ? (
          <LiveIndicator label={`Élő · ${standings.live.label}`} />
        ) : (
          <span className="tabular font-mono text-sm text-muted-foreground">
            <span className="font-semibold text-foreground">{doneCount}</span>{" "}
            / {legs.length} forduló
          </span>
        )
      }
    >
      <div className="flex gap-1" aria-hidden="true">
        {legs.map((leg, index) => (
          <span
            key={standings.competitions[index].competitionWcaId}
            className={`h-1.5 flex-1 rounded-full ${LEG_CLASSES[leg]}`}
          />
        ))}
      </div>

      <ol className="mt-6 flex flex-col divide-y divide-line">
        {top.map((person, index) => (
          <li
            key={person.wcaId}
            className={`items-center gap-4 py-2.5 first:pt-0 last:pb-0 ${
              index < COMPACT_COUNT ? "flex" : "hidden lg:flex"
            }`}
          >
            <span
              className={`tabular w-6 shrink-0 font-mono text-sm font-semibold ${
                index === 0 ? "text-brand-red" : "text-subtle-foreground"
              }`}
            >
              {index + 1}.
            </span>
            <span
              className={`min-w-0 flex-1 truncate ${
                index === 0
                  ? "font-semibold text-foreground"
                  : "text-muted-foreground"
              }`}
            >
              {person.name}
            </span>
            <span className="tabular shrink-0 font-mono text-sm font-semibold text-foreground">
              {person.totalPoints}
              <span className="ml-1 text-xs font-normal text-subtle-foreground">
                p
              </span>
            </span>
          </li>
        ))}
      </ol>

      <div className="mt-auto flex flex-col gap-2 pt-5 text-xs leading-relaxed text-subtle-foreground">
        {standings.people.length > COMPACT_COUNT && (
          <p>
            Összesen{" "}
            <span className="tabular font-mono">{standings.people.length}</span>{" "}
            versenyző a ranglistán
          </p>
        )}
        {standings.live && (
          <p>
            Az élő forduló pontjai ideiglenesek, a hivatalos eredmények
            feltöltésekor véglegesednek.
          </p>
        )}
      </div>
    </Tile>
  );
}
