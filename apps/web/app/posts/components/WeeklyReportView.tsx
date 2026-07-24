import {
  WeeklyReport,
  WeeklyReportCompetition,
  WeeklyReportResult,
} from "../types";
import { flagEmoji, formatPeriod } from "../lib/report";

/**
 * A heti riport strukturált megjelenítése (a `tag === "report"` posztokhoz).
 *
 * A tartalom NEM markdown, hanem JSON (lásd lib/report.ts + a weekly-report job).
 * Így a megjelenés teljesen itt, komponensben stílusozható — az email riporthoz
 * hasonló, szekciós/kártyás elrendezés, tetszőleges kontrollal.
 */

const badgeColors: Record<string, string> = {
  WR: "bg-yellow-400 text-yellow-950",
  CR: "bg-blue-500 text-white",
  NR: "bg-green-600 text-white",
  PR: "bg-gray-400 text-gray-950",
};

function ResultChip({ result }: { result: WeeklyReportResult }) {
  const badge = badgeColors[result.badge] ?? "bg-gray-300 text-gray-900";
  return (
    <span className="inline-flex items-center gap-1.5 rounded-md bg-gray-100 px-2 py-1 text-sm dark:bg-gray-800">
      <span className="font-medium">{result.event}</span>
      <span className="text-gray-500 dark:text-gray-400">{result.type}</span>
      <span className="font-mono">{result.result}</span>
      <span className={`rounded px-1.5 py-0.5 text-xs font-bold ${badge}`}>
        {result.badge}
      </span>
    </span>
  );
}

function CompetitionCard({ competition }: { competition: WeeklyReportCompetition }) {
  return (
    <div className="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 border-b border-gray-200 bg-gray-50 px-4 py-3 dark:border-gray-700 dark:bg-gray-800/50">
        <h3 className="text-lg font-semibold">
          {competition.name}{" "}
          <span aria-hidden>{flagEmoji(competition.countryIso)}</span>
        </h3>
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {formatPeriod(competition.from, competition.to)}
        </span>
      </div>

      <ul className="divide-y divide-gray-100 dark:divide-gray-800">
        {competition.competitors.map((competitor, i) => (
          <li
            key={`${competitor.name}-${i}`}
            className="flex flex-wrap items-center gap-x-3 gap-y-2 px-4 py-2.5"
          >
            <span className="font-medium">{competitor.name}</span>
            {competitor.results.length > 0 ? (
              <span className="flex flex-wrap gap-1.5">
                {competitor.results.map((result, j) => (
                  <ResultChip key={j} result={result} />
                ))}
              </span>
            ) : (
              <span className="text-sm text-gray-400 dark:text-gray-500">
                résztvevő
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="mb-4 mt-10 border-b border-gray-200 pb-1 text-2xl font-bold first:mt-0 dark:border-gray-700">
      {children}
    </h2>
  );
}

export default function WeeklyReportView({ report }: { report: WeeklyReport }) {
  const { competitions, competitorMilestones, delegateMilestones } = report;
  const isEmpty =
    competitions.length === 0 &&
    competitorMilestones.length === 0 &&
    delegateMilestones.length === 0;

  if (isEmpty) {
    return (
      <p className="text-gray-500 dark:text-gray-400">
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
            <CompetitionCard key={`${competition.name}-${i}`} competition={competition} />
          ))}
        </div>
      ) : (
        <p className="text-gray-500 dark:text-gray-400">
          Nem volt ezen a héten magyar érdekeltségű versenyeredmény.
        </p>
      )}

      {competitorMilestones.length > 0 && (
        <>
          <SectionTitle>Versenyzői mérföldkövek</SectionTitle>
          <ul className="space-y-2">
            {competitorMilestones.map((milestone, i) => (
              <li key={i} className="flex flex-wrap items-baseline gap-x-2">
                <span className="font-medium">{milestone.name}</span>
                <span className="text-gray-500 dark:text-gray-400">
                  — {milestone.count} verseny
                </span>
                {milestone.estimated && (
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    ({milestone.achieved ? "teljesítve" : "tervezve"}:{" "}
                    {milestone.estimated}
                    {milestone.estimatedDate ? `, ${milestone.estimatedDate}` : ""})
                  </span>
                )}
              </li>
            ))}
          </ul>
        </>
      )}

      {delegateMilestones.length > 0 && (
        <>
          <SectionTitle>Delegálti mérföldkövek</SectionTitle>
          <ul className="space-y-2">
            {delegateMilestones.map((milestone, i) => (
              <li key={i} className="flex flex-wrap items-baseline gap-x-2">
                <span className="font-medium">{milestone.name}</span>
                <span className="text-gray-500 dark:text-gray-400">
                  — {milestone.count} verseny
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}
