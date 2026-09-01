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
  "sticky top-0 border-b border-line bg-surface-muted px-3 py-2.5 " +
  "text-xs font-semibold uppercase tracking-wide text-muted-foreground";

// Helyezés + név + pont EGY cellában, és ez a cella marad a helyén vízszintes
// görgetéskor. Egy 12 állomásos évadban enélkül nem lehetne megmondani, kinek
// az eredményeit nézzük — és épp a végösszeg csúszna ki elsőként a képből.
// Azért egyetlen cella, mert több ragadós oszlopnál a `left` eltolásoknak
// pixelre egyezniük kellene az oszlopszélességekkel.
const FROZEN =
  "sticky left-0 border-r border-line " +
  "min-w-[13.5rem] max-w-[17rem] sm:min-w-[15rem]";

const BODY_CELL = "border-b border-line px-3 py-2";
const ROW_HOVER =
  "transition-colors duration-200 ease-smooth group-hover:bg-surface-muted";

const rankChipClass = (isPodium: boolean) =>
  "tabular inline-flex min-w-7 shrink-0 justify-center rounded-full px-2 py-0.5 font-mono text-xs font-semibold " +
  (isPodium
    ? "bg-brand-green/15 text-brand-green ring-1 ring-inset ring-brand-green/30"
    : "text-muted-foreground");

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
    <div className="max-h-[75vh] overflow-auto rounded-card border border-line bg-surface">
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
            const byCompetition = groupByCompetition(person.results);

            return (
              <tr key={person.wcaId} className="group">
                <th
                  scope="row"
                  className={`${BODY_CELL} ${FROZEN} ${ROW_HOVER} z-10 bg-surface text-left font-normal`}
                >
                  <div className="flex items-center gap-2.5">
                    <span className={rankChipClass(rank <= podiumCount)}>
                      {rank}
                    </span>
                    <span className="min-w-0 flex-1 truncate font-medium text-foreground">
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
