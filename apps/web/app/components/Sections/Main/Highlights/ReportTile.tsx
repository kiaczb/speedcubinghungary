import { formatDayMonth } from "@/app/lib/date";
import { eventName } from "@/app/lib/events";
import {
  RECORD_BADGE_COLORS,
  RECORD_BADGE_FALLBACK,
} from "@/app/posts/lib/recordBadges";
import { PostsItem, WeeklyReport } from "@/app/posts/types";
import { Tile } from "./Tile";

const HIGHLIGHT_COUNT = 3;
const RECORD_RANK = ["WR", "CR", "NR"];

function records(report: WeeklyReport) {
  return report.competitions
    .flatMap((competition) =>
      competition.competitors.flatMap((competitor) =>
        competitor.results.map((result) => ({ name: competitor.name, result })),
      ),
    )
    .filter(({ result }) => RECORD_RANK.includes(result.badge))
    .sort(
      (a, b) =>
        RECORD_RANK.indexOf(a.result.badge) -
        RECORD_RANK.indexOf(b.result.badge),
    );
}

export const ReportTile = ({
  post,
  report,
  className = "",
}: {
  post: PostsItem;
  report: WeeklyReport;
  className?: string;
}) => {
  const all = records(report);
  const top = all.slice(0, HIGHLIGHT_COUNT);
  const milestone = report.competitorMilestones.find((m) => m.achieved);

  return (
    <Tile
      href={`/posts/${post.id}`}
      title="A hét eredményei"
      meta={
        <span className="tabular font-mono text-xs text-subtle-foreground">
          {formatDayMonth(post.created_at)}
        </span>
      }
      cta="Teljes riport"
      className={className}
    >
      <p className="tabular font-mono text-xs text-muted-foreground">
        {report.competitions.length} verseny · {all.length} rekord
      </p>

      {top.length > 0 ? (
        <ul className="mt-4 flex flex-col gap-2.5">
          {top.map(({ name, result }, index) => (
            <li
              key={`${name}-${result.event}-${result.type}-${index}`}
              className="flex items-center gap-2 text-sm"
            >
              <span
                className={`shrink-0 rounded px-1.5 py-0.5 font-mono text-[10px] font-bold ${
                  RECORD_BADGE_COLORS[result.badge] ?? RECORD_BADGE_FALLBACK
                }`}
              >
                {result.badge}
              </span>
              <span className="min-w-0 flex-1 truncate text-foreground">
                {name}
              </span>
              <span className="tabular shrink-0 font-mono text-xs text-muted-foreground">
                {eventName(result.event)}{" "}
                {result.type === "average" ? "átlag" : "single"}{" "}
                <span className="text-foreground">{result.result}</span>
              </span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
          Ezen a héten nem született rekord.
        </p>
      )}

      {milestone && (
        <p className="mt-auto pt-4 text-xs text-subtle-foreground">
          Mérföldkő: {milestone.name} –{" "}
          <span className="tabular font-mono">{milestone.count}.</span> verseny
        </p>
      )}
    </Tile>
  );
};
