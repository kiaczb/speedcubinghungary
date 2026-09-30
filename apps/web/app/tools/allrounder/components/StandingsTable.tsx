import EventChips from "./EventChips";
import { groupByCompetition, medalAccentClass, medalForRank, medalTextClass } from "../lib/results";
import type { PersonStandings, SeriesCompetition } from "../types";

type StandingsTableProps = {
  competitions: SeriesCompetition[];
  people: PersonStandings[];
  podiumCount: number;
};

const HEAD_CELL =
  "sticky top-0 border-b border-line bg-surface-muted px-3 py-2.5 " +
  "text-xs font-semibold uppercase tracking-wide text-muted-foreground";

const FROZEN =
  "sticky left-0 border-r border-line " +
  "w-56 min-w-56 max-w-56";

const BODY_CELL = "border-b border-line px-3 py-2";
const ROW_HOVER =
  "transition-colors duration-200 ease-smooth group-hover:bg-surface-muted";

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
            const medal = medalForRank(rank);
            const byCompetition = groupByCompetition(person.results);

            return (
              <tr key={person.wcaId} className="group">
                <th
                  scope="row"
                  className={`${BODY_CELL} ${FROZEN} ${ROW_HOVER} relative z-10 bg-surface text-left font-normal`}
                >
                  {isPodium && (
                    <span
                      className={`absolute inset-y-0 left-0 w-0.5 ${
                        medal ? medalAccentClass(medal) : "bg-brand-green"
                      }`}
                      aria-hidden
                    />
                  )}

                  <div className="flex items-center gap-2.5">
                    <span
                      className={`tabular w-6 shrink-0 text-right font-mono text-xs ${
                        medal
                          ? `font-semibold ${medalTextClass(medal)}`
                          : isPodium
                            ? "font-semibold text-foreground"
                            : "text-muted-foreground"
                      }`}
                    >
                      {rank}
                    </span>
                    <span
                      className="min-w-0 flex-1 truncate text-sm font-medium text-foreground"
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
