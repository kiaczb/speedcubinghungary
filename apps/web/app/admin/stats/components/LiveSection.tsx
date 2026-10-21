"use client";

import { eventName } from "@/app/lib/events";
import { formatTimestamp } from "@/app/lib/date";
import type { LiveStats, Phase } from "../types";
import { formatNumber, formatPercent, resultType } from "../lib/format";
import {
  KeyFigures,
  PersonName,
  RecordBadge,
  Rows,
  StatBlock,
  StatSection,
} from "./Blocks";

export default function LiveSection({
  live,
  error,
  phase,
}: {
  live?: LiveStats;
  error?: string;
  phase: Phase;
}) {
  const title = phase === "after" ? "WCA Live" : "Élőben (WCA Live)";

  if (!live) {
    return (
      <StatSection title={title}>
        <p className="text-sm text-muted-foreground">{error}</p>
      </StatSection>
    );
  }

  return (
    <StatSection
      title={title}
      description={
        <>
          A beírt eredmények alapján, percenként frissítve. Utolsó lekérés:{" "}
          <span className="tabular font-mono">{formatTimestamp(live.fetchedAt)}</span>
          {phase === "after" &&
            " — a PR-eket csak a WCA Live jelöli, a hivatalos eredmények nem."}
        </>
      }
    >
      <KeyFigures
        figures={[
          {
            label: "Lezárt fordulók",
            value: `${live.roundsFinished}/${live.roundsTotal}`,
            sub: `${live.roundsOpen} megnyitva`,
          },
          { label: "Beírt eredmények", value: formatNumber(live.resultsEntered) },
          { label: "Kirakások", value: formatNumber(live.attempts) },
          {
            label: "DNF",
            value: live.dnfs,
            sub: formatPercent(live.dnfs, live.attempts),
          },
          { label: "PR-ek", value: live.prs },
          { label: "Rekordok", value: live.records.length },
        ]}
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <StatBlock title="Rekordok">
          <Rows
            empty="Még nem dőlt meg rekord."
            rows={live.records.map((record, index) => ({
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

        <StatBlock title="Legtöbb PR">
          <Rows
            empty="Még nincs PR."
            rows={live.prLeaders.map((leader) => ({
              key: leader.person.wcaId ?? leader.person.name,
              label: <PersonName person={leader.person} />,
              value: `${leader.value} PR`,
            }))}
          />
        </StatBlock>

        <StatBlock
          title="Versenyszámok legjobbjai eddig"
          note="Az összes forduló legjobb single-je és átlaga."
          className="lg:col-span-2"
        >
          <Rows
            limit={20}
            empty="Még nincs beírt eredmény."
            rows={live.eventBests.map((best) => ({
              key: best.eventId,
              label: <span className="text-foreground">{eventName(best.eventId)}</span>,
              sub: [
                best.single && `single: ${best.single.person.name}`,
                best.average && `átlag: ${best.average.person.name}`,
              ]
                .filter(Boolean)
                .join(" · "),
              value: [best.single?.display, best.average?.display]
                .filter(Boolean)
                .join(" / "),
            }))}
          />
        </StatBlock>
      </div>
    </StatSection>
  );
}
