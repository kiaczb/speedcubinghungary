import Badge from "../../../Badge";
import { formatDateTime } from "@/app/lib/date";
import {
  PredictionContest,
  statusBadgeMap,
} from "@/app/tools/prediction/types";
import { Tile } from "./Tile";

const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

function timeLeft(deadline: string, now: Date): string {
  const remaining = new Date(deadline).getTime() - now.getTime();
  if (remaining >= DAY) return `még ${Math.floor(remaining / DAY)} nap`;
  if (remaining >= HOUR) return `még ${Math.floor(remaining / HOUR)} óra`;
  return "hamarosan zárul";
}

export function nearestOpenContest(
  contests: PredictionContest[],
): PredictionContest | undefined {
  return contests
    .filter((contest) => contest.status === "open")
    .sort((a, b) =>
      a.predictionDeadline.localeCompare(b.predictionDeadline),
    )[0];
}

export function latestContest(
  contests: PredictionContest[],
): PredictionContest | undefined {
  return [...contests].sort((a, b) =>
    b.competitionStartDate.localeCompare(a.competitionStartDate),
  )[0];
}

export const OpenPredictionTile = ({
  contest,
  now,
  className = "",
}: {
  contest: PredictionContest;
  now: Date;
  className?: string;
}) => (
  <Tile
    href={`/tools/prediction/${contest.id}`}
    title="Tippverseny"
    meta={
      <Badge
        label={statusBadgeMap.open.label}
        color={statusBadgeMap.open.color}
      />
    }
    cta="Tippelek"
    className={className}
  >
    <p className="font-medium leading-snug text-foreground">
      {contest.competitionName}
    </p>
    <p className="mt-1 text-sm text-muted-foreground">
      Ki áll a dobogón? Add le a tipped a határidőig.
    </p>
    <p className="mt-auto pt-4 text-sm">
      <span className="font-semibold text-brand-red">
        {timeLeft(contest.predictionDeadline, now)}
      </span>
      <span className="tabular font-mono text-xs text-subtle-foreground">
        {" "}
        · {formatDateTime(contest.predictionDeadline)}
      </span>
    </p>
  </Tile>
);

export const PredictionFallbackTile = ({
  latest,
  className = "",
}: {
  latest?: PredictionContest;
  className?: string;
}) => (
  <Tile
    href="/tools/prediction"
    title="Tippverseny"
    cta={latest ? "Korábbi tippversenyek" : "Megnyitás"}
    className={className}
  >
    <p className="text-sm leading-relaxed text-muted-foreground">
      Most nincs nyitott tippverseny. A közelgő versenyek előtt itt tippelhetsz
      a dobogósokra.
    </p>
    {latest && (
      <p className="mt-auto pt-4 text-xs text-subtle-foreground">
        Legutóbbi: {latest.competitionName}
      </p>
    )}
  </Tile>
);
