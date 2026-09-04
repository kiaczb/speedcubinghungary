import EventChips from "./EventChips";
import type { EventResult, PersonStandings, SeriesCompetition } from "../types";

type StandingsTableProps = {
  competitions: SeriesCompetition[];
  people: PersonStandings[];
  podiumCount: number;
};

// A táblázat `border-separate`, nem `border-collapse`: összevont keretnél a
// böngészők nem viszik magukkal a szegélyt a ragadós (sticky) cellákkal, így
// vízszintes görgetéskor szétesne a rács.
const HEAD_CELL =
  "sticky top-0 border-b border-line bg-surface-muted px-2.5 py-2.5 sm:px-3 " +
  "text-xs font-semibold uppercase tracking-wide text-muted-foreground";


const FROZEN =
  "sticky left-0 border-r border-line " +
  "w-44 min-w-44 max-w-44 sm:w-56 sm:min-w-56 sm:max-w-56";

const BODY_CELL = "border-b border-line px-2.5 py-2 sm:px-3";
const ROW_HOVER =
  "transition-colors duration-200 ease-smooth group-hover:bg-surface-muted";

/** Egy versenyző beszámító eredményei állomásonként csoportosítva. */
function groupByCompetition(results: EventResult[]): Map<string, EventResult[]> {
  const grouped = new Map<string, EventResult[]>();

  for (const result of results) {
    const existing = grouped.get(result.competitionWcaId);
    if (existing) {
      existing.push(result);
    } else {
      grouped.set(result.competitionWcaId, [result]);
    }
  }

  return grouped;
}

const StandingsTable = ({
  competitions,
  people,
  podiumCount,
}: StandingsTableProps) => {
  return (
   
    <div className="relative max-h-[75vh] overflow-auto overscroll-contain rounded-card border border-line bg-surface">
      <table className="min-w-full border-separate border-spacing-0 text-sm">
        <thead>
          <tr>
            <th scope="col" className={`${HEAD_CELL} ${FROZEN} z-30`}>
              <div className="flex items-baseline justify-between gap-3">
                <span>Versenyző</span>
                <span>Pont</span>
              </div>
            </th>

            {competitions.map((competition) => (
              <th
                key={competition.competitionWcaId}
                scope="col"
                className={`${HEAD_CELL} z-20 whitespace-nowrap text-center`}
              >
                {competition.label}
              </th>
            ))}
          </tr>
        </thead>

        <tbody>
          {people.map((person, index) => {
            const rank = index + 1;
            const isPodium = rank <= podiumCount;
            const byCompetition = groupByCompetition(person.results);

            return (
              <tr key={person.wcaId} className="group">
                <th
                  scope="row"
                  className={`${BODY_CELL} ${FROZEN} ${ROW_HOVER} relative z-10 bg-surface text-left font-normal`}
                >
                  {/* A díjazott helyezéseket a sor bal szélén futó vonal jelöli:
                      abszolút pozícióban, hogy ne tolja el a cella tartalmát. */}
                  {isPodium && (
                    <span
                      className="absolute inset-y-0 left-0 w-0.5 bg-brand-green"
                      aria-hidden
                    />
                  )}

                  <div className="flex items-center gap-2 sm:gap-2.5">
                    <span
                      className={`tabular w-5 shrink-0 text-right font-mono text-xs sm:w-6 ${
                        isPodium
                          ? "font-semibold text-foreground"
                          : "text-muted-foreground"
                      }`}
                    >
                      {rank}
                    </span>
                    <span
                      className="min-w-0 flex-1 truncate text-[13px] font-medium text-foreground sm:text-sm"
                      title={person.name}
                    >
                      {person.name}
                    </span>
                    <span className="tabular shrink-0 font-mono font-semibold text-foreground">
                      {person.totalPoints}
                    </span>
                  </div>
                </th>

                {competitions.map((competition) => (
                  <td
                    key={competition.competitionWcaId}
                    className={`${BODY_CELL} ${ROW_HOVER}`}
                  >
                    <EventChips
                      results={
                        byCompetition.get(competition.competitionWcaId) ?? []
                      }
                    />
                  </td>
                ))}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

export default StandingsTable;
