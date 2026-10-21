"use client";

import { eventName } from "@/app/lib/events";
import type { ResultStats } from "../types";
import {
  countryName,
  formatNumber,
  formatPercent,
  formatSolvingTime,
  resultType,
} from "../lib/format";
import {
  KeyFigures,
  PersonName,
  RecordBadge,
  Rows,
  StatBlock,
  StatSection,
} from "./Blocks";

export default function ResultsSection({
  results,
  error,
}: {
  results?: ResultStats;
  error?: string;
}) {
  if (!results) {
    return (
      <StatSection title="Hivatalos eredmények">
        <p className="text-sm text-muted-foreground">{error}</p>
      </StatSection>
    );
  }

  return (
    <StatSection
      title="Hivatalos eredmények"
      description="A WCA-ra feltöltött, végleges eredmények alapján."
    >
      <KeyFigures
        figures={[
          { label: "Kirakások", value: formatNumber(results.solves) },
          {
            label: "DNF",
            value: results.dnfs,
            sub: formatPercent(results.dnfs, results.solves),
          },
          {
            label: "Összes kirakási idő",
            value: formatSolvingTime(results.solvingCentiseconds),
            sub: "FMC és MBLD nélkül",
          },
          { label: "Rekordok", value: results.records.length },
          { label: "Sub-10 3x3 single", value: results.sub10Singles.length, sub: "versenyző" },
          { label: "Versenyszámok", value: results.winners.length },
        ]}
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <StatBlock title="Győztesek" className="lg:row-span-2">
          <Rows
            limit={20}
            empty="Nincs döntő eredmény."
            rows={results.winners.map((winner) => ({
              key: winner.eventId,
              label: <PersonName person={winner.person} />,
              sub: eventName(winner.eventId),
              value: [winner.best, winner.average].filter(Boolean).join(" / "),
            }))}
          />
        </StatBlock>

        <StatBlock title="Rekordok">
          <Rows
            empty="Nem dőlt meg rekord."
            rows={results.records.map((record, index) => ({
              key: `${record.person.wcaId}-${record.eventId}-${record.type}-${index}`,
              label: (
                <span>
                  <PersonName person={record.person} />
                  <RecordBadge record={record.record} />
                </span>
              ),
              sub: `${eventName(record.eventId)} ${resultType(record.type)}`,
              value: record.display,
            }))}
          />
        </StatBlock>

        <StatBlock title="Legtöbb érem" note="arany · ezüst · bronz, a döntőkből">
          <Rows
            empty="Nincs érem."
            rows={results.podiums.map((holder) => ({
              key: holder.person.wcaId ?? holder.person.name,
              label: <PersonName person={holder.person} />,
              value: `${holder.gold} · ${holder.silver} · ${holder.bronze}`,
            }))}
          />
        </StatBlock>

        <StatBlock title="Éremtáblázat országonként" note="arany · ezüst · bronz">
          <Rows
            empty="Nincs érem."
            rows={results.countryPodiums.map((country) => ({
              key: country.countryIso2,
              label: <span className="text-foreground">{countryName(country.countryIso2)}</span>,
              value: `${country.gold} · ${country.silver} · ${country.bronze}`,
            }))}
          />
        </StatBlock>

        {results.sub10Singles.length > 0 && (
          <StatBlock title="Sub-10 3x3 single">
            <p className="text-sm leading-relaxed text-foreground">
              {results.sub10Singles.map((person) => person.name).join(", ")}
            </p>
          </StatBlock>
        )}
      </div>
    </StatSection>
  );
}
