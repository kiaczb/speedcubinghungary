import {
  WeeklyReport,
  WeeklyReportCompetition,
  WeeklyReportResult,
} from "../types";
import { flagEmoji, formatPeriod } from "../lib/report";
import { RECORD_BADGE_COLORS, RECORD_BADGE_FALLBACK } from "../lib/recordBadges";

function ResultChip({ result }: { result: WeeklyReportResult }) {
  const badge = RECORD_BADGE_COLORS[result.badge] ?? RECORD_BADGE_FALLBACK;
  return (
    <span className="inline-flex items-center gap-1.5 rounded-control border border-line bg-surface px-2 py-1 text-sm">
      <span className="font-medium text-foreground">{result.event}</span>
      <span className="text-subtle-foreground">{result.type}</span>
      <span className="tabular font-mono text-foreground">{result.result}</span>
      <span
        className={`rounded px-1.5 py-0.5 font-mono text-xs font-bold ${badge}`}
      >
        {result.badge}
      </span>
    </span>
  );
}

function CompetitionCard({
  competition,
}: {
  competition: WeeklyReportCompetition;
}) {
  return (
    <div className="overflow-hidden rounded-card border border-line">
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 border-b border-line bg-surface-muted px-4 py-3">
        <h3 className="font-semibold text-foreground">
          {competition.name} <span aria-hidden>{flagEmoji(competition.countryIso)}</span>
        </h3>
        <span className="tabular font-mono text-xs text-muted-foreground">
          {formatPeriod(competition.from, competition.to)}
        </span>
      </div>

      <ul className="divide-y divide-line">
        {competition.competitors.map((competitor, i) => (
          <li
            key={`${competitor.name}-${i}`}
            className="flex flex-wrap items-center gap-x-3 gap-y-2 px-4 py-3"
          >
            <span className="font-medium text-foreground">
              {competitor.name}
            </span>
            {competitor.results.length > 0 ? (
              <span className="flex flex-wrap gap-1.5">
                {competitor.results.map((result, j) => (
                  <ResultChip key={j} result={result} />
                ))}
              </span>
            ) : (
              <span className="text-sm text-subtle-foreground">résztvevő</span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="mb-4 mt-10 border-b border-line pb-2 text-xl font-bold text-foreground first:mt-0">
      {children}
    </h2>
  );
}

function MilestoneList({ children }: { children: React.ReactNode }) {
  return <ul className="space-y-2.5">{children}</ul>;
}

export default function WeeklyReportView({ report }: { report: WeeklyReport }) {
  const { competitions, competitorMilestones, delegateMilestones } = report;
  const isEmpty =
    competitions.length === 0 &&
    competitorMilestones.length === 0 &&
    delegateMilestones.length === 0;

  if (isEmpty) {
    return (
      <p className="text-muted-foreground">
        Ezen a héten nem volt magyar érdekeltségű esemény.
      </p>
    );
  }

  return (
    <div>
      <SectionTitle>Eheti versenyeredmények</SectionTitle>
      {competitions.length > 0 ? (
        <div className="space-y-5">
          {competitions.map((competition, i) => (
            <CompetitionCard
              key={`${competition.name}-${i}`}
              competition={competition}
            />
          ))}
        </div>
      ) : (
        <p className="text-muted-foreground">
          Nem volt ezen a héten magyar érdekeltségű versenyeredmény.
        </p>
      )}

      {competitorMilestones.length > 0 && (
        <>
          <SectionTitle>Versenyzői mérföldkövek</SectionTitle>
          <MilestoneList>
            {competitorMilestones.map((milestone, i) => (
              <li key={i} className="flex flex-wrap items-baseline gap-x-2">
                <span className="font-medium text-foreground">
                  {milestone.name}
                </span>
                <span className="text-muted-foreground">
                  — <span className="tabular font-mono">{milestone.count}</span>{" "}
                  verseny
                </span>
                {milestone.estimated && (
                  <span className="text-sm text-subtle-foreground">
                    ({milestone.achieved ? "teljesítve" : "tervezve"}:{" "}
                    {milestone.estimated}
                    {milestone.estimatedDate ? `, ${milestone.estimatedDate}` : ""})
                  </span>
                )}
              </li>
            ))}
          </MilestoneList>
        </>
      )}

      {delegateMilestones.length > 0 && (
        <>
          <SectionTitle>Delegálti mérföldkövek</SectionTitle>
          <MilestoneList>
            {delegateMilestones.map((milestone, i) => (
              <li key={i} className="flex flex-wrap items-baseline gap-x-2">
                <span className="font-medium text-foreground">
                  {milestone.name}
                </span>
                <span className="text-muted-foreground">
                  — <span className="tabular font-mono">{milestone.count}</span>{" "}
                  verseny
                </span>
              </li>
            ))}
          </MilestoneList>
        </>
      )}
    </div>
  );
}
