"use client";

import "@cubing/icons";
import { eventName } from "@/app/lib/events";
import type { CompetitionInfo, FieldStats } from "../types";
import {
  countryName,
  formatNumber,
  formatPercent,
  genderName,
  resultType,
} from "../lib/format";
import {
  BarList,
  KeyFigures,
  PersonName,
  RecordBadge,
  Rows,
  StatBlock,
  StatSection,
} from "./Blocks";

const EventLabel = ({ eventId }: { eventId: string }) => (
  <span className="inline-flex items-center gap-1.5">
    <span
      aria-hidden
      className={`cubing-icon event-${eventId} text-base leading-none text-muted-foreground`}
    />
    {eventName(eventId)}
  </span>
);

export default function FieldSection({
  competition,
  field,
  error,
}: {
  competition: CompetitionInfo;
  field?: FieldStats;
  error?: string;
}) {
  if (!field) {
    return (
      <StatSection title="Mezőny">
        <p className="text-sm text-muted-foreground">{error}</p>
      </StatSection>
    );
  }

  const women = field.genders.find((gender) => gender.key === "f")?.count ?? 0;

  return (
    <StatSection
      title="Mezőny"
      description="Az elfogadott, versenyző nevezések alapján (WCIF), kiegészítve a versenyzők WCA-összesítőjével."
    >
      <KeyFigures
        figures={[
          {
            label: "Versenyzők",
            value: field.limit
              ? `${field.competitors}/${field.limit}`
              : field.competitors,
          },
          {
            label: "Újoncok",
            value: field.newcomers,
            sub: formatPercent(field.newcomers, field.competitors),
          },
          {
            label: "Külföldiek",
            value: field.foreigners,
            sub: `${field.countries.length} ország`,
          },
          {
            label: "Nők",
            value: women,
            sub: formatPercent(women, field.competitors),
          },
          {
            label: "Versenyszám / fő",
            value: field.averageEvents.toLocaleString("hu-HU"),
          },
          {
            label: "Eddigi kirakásaik",
            value: formatNumber(field.careerSolves),
            sub: "a mezőny összesen",
          },
        ]}
      />

      {field.summariesError && (
        <p className="text-sm text-muted-foreground">{field.summariesError}</p>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <StatBlock title="Nevezések versenyszámonként">
          <BarList
            limit={20}
            items={field.events.map((event) => ({
              key: event.key,
              label: <EventLabel eventId={event.key} />,
              count: event.count,
            }))}
          />
        </StatBlock>

        <StatBlock
          title="Országok"
          note={field.genders
            .map((gender) => `${genderName(gender.key)}: ${gender.count}`)
            .join(" · ")}
        >
          <BarList
            limit={8}
            items={field.countries.map((country) => ({
              key: country.key,
              label: countryName(country.key),
              count: country.count,
            }))}
          />
        </StatBlock>

        <StatBlock
          title="Mérföldkövek"
          note="Kinek hányadik versenye ez, a WCA versenyszámlálója alapján."
        >
          <Rows
            empty="Senkinek sem kerek számú versenye ez."
            rows={field.milestones.map((milestone) => ({
              key: milestone.person.wcaId ?? milestone.person.name,
              label: <PersonName person={milestone.person} />,
              value: `${milestone.value}. verseny`,
            }))}
          />
        </StatBlock>

        <StatBlock
          title="Évfordulósok"
          note="A WCA ID-ben szereplő év az első verseny éve."
        >
          <Rows
            empty="Nincs kerek évfordulós versenyző."
            rows={field.anniversaries.map((anniversary) => ({
              key: String(anniversary.years),
              label: anniversary.people.map((p) => p.name).join(", "),
              sub: `${anniversary.startYear} óta`,
              value: `${anniversary.years} éve`,
            }))}
          />
        </StatBlock>

        <StatBlock title="Legrégebben versenyzők">
          <Rows
            empty="Csak újoncok neveztek."
            rows={field.veterans.map((veteran) => ({
              key: veteran.person.wcaId ?? veteran.person.name,
              label: <PersonName person={veteran.person} />,
              value: `${veteran.value} óta`,
            }))}
          />
        </StatBlock>

        <StatBlock title="Legtöbb verseny">
          <Rows
            empty="Nincs adat."
            rows={field.mostCompetitions.map((entry) => ({
              key: entry.person.wcaId ?? entry.person.name,
              label: <PersonName person={entry.person} />,
              value: `${entry.value} verseny`,
            }))}
          />
        </StatBlock>

        <StatBlock title="Legjobb világranglista-helyezések">
          <Rows
            empty="Nincs rangsorolt versenyző."
            rows={field.topRanked.map((ranked) => ({
              key: `${ranked.person.wcaId}-${ranked.eventId}-${ranked.type}`,
              label: <PersonName person={ranked.person} />,
              sub: `${eventName(ranked.eventId)} ${resultType(ranked.type)} · ${ranked.display}`,
              value: `#${ranked.worldRanking}`,
            }))}
          />
        </StatBlock>

        <StatBlock
          title="Országos csúcstartók a mezőnyben"
          note="Saját országuk aktuális rekordere — ezekben a számokban dőlhet meg rekord."
        >
          <Rows
            empty="Nincs országos csúcstartó a nevezők között."
            rows={field.recordHolders.map((holder) => ({
              key: `${holder.person.wcaId}-${holder.eventId}-${holder.type}`,
              label: (
                <span>
                  <PersonName person={holder.person} />
                  {holder.record && <RecordBadge record={holder.record} />}
                </span>
              ),
              sub: `${eventName(holder.eventId)} ${resultType(holder.type)}`,
              value: holder.display,
            }))}
          />
        </StatBlock>

        <StatBlock title="Legtöbb érem" note="arany · ezüst · bronz, a teljes pályafutásból">
          <Rows
            empty="Még egyik nevezőnek sincs WCA-érme."
            rows={field.mostMedals.map((holder) => ({
              key: holder.person.wcaId ?? holder.person.name,
              label: <PersonName person={holder.person} />,
              value: `${holder.gold} · ${holder.silver} · ${holder.bronze}`,
            }))}
          />
        </StatBlock>

        <StatBlock title="Legtöbb rekord">
          <Rows
            empty="Egyik nevezőnek sincs rekordja."
            rows={field.mostRecords.map((holder) => ({
              key: holder.person.wcaId ?? holder.person.name,
              label: <PersonName person={holder.person} />,
              value: [
                holder.world && `${holder.world} WR`,
                holder.continental && `${holder.continental} CR`,
                holder.national && `${holder.national} NR`,
              ]
                .filter(Boolean)
                .join(" · "),
            }))}
          />
        </StatBlock>

        {field.allEvents.length > 0 && (
          <StatBlock
            title="Mindenre neveztek"
            note={`Mind a(z) ${competition.eventIds.length} versenyszámra.`}
          >
            <p className="text-sm leading-relaxed text-foreground">
              {field.allEvents.map((person) => person.name).join(", ")}
            </p>
          </StatBlock>
        )}
      </div>
    </StatSection>
  );
}
