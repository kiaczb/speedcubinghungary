"use client";

import { formatLongDate } from "@/app/lib/date";
import type { CompetitionInfo, HistoryStats } from "../types";
import { countryName, formatGap, inYear } from "../lib/format";
import { PersonName, Rows, StatBlock, StatSection } from "./Blocks";

export default function HistorySection({
  competition,
  history,
}: {
  competition: CompetitionInfo;
  history: HistoryStats;
}) {
  if (history.status === "unavailable") return null;

  const description =
    "A versenyzők korábbi versenyei alapján. Ezt a WCA csak versenyzőnként adja ki (fejenként ~1 mp), ezért a háttérben gyűlik, és egy napig tároljuk.";

  if (history.status === "running") {
    const percent = history.total > 0 ? (history.done / history.total) * 100 : 0;
    return (
      <StatSection title="Visszatérők és pályafutás" description={description}>
        <div className="max-w-md">
          <div className="flex items-baseline justify-between text-sm">
            <span className="text-muted-foreground">Versenytörténetek lekérése</span>
            <span className="tabular font-mono text-foreground">
              {history.done}/{history.total}
            </span>
          </div>
          <div className="mt-2 h-2 rounded-full bg-surface-muted">
            <div
              className="h-2 rounded-full bg-brand-green transition-[width] duration-500 ease-smooth"
              style={{ width: `${percent}%` }}
            />
          </div>
          <p className="mt-2 text-xs text-subtle-foreground">
            Az oldal magától frissül, amíg tart.
          </p>
        </div>
      </StatSection>
    );
  }

  const country = countryName(competition.countryIso2);

  return (
    <StatSection title="Visszatérők és pályafutás" description={description}>
      {history.failed > 0 && (
        <p className="text-sm text-muted-foreground">
          {history.failed} versenyző versenytörténete nem jött le, ők kimaradtak.
          Néhány perc múlva újrapróbáljuk.
        </p>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <StatBlock
          title="Visszatérők"
          note="Akik legalább egy éve nem versenyeztek — a legutóbbi versenyükhöz képest, nem az elsőhöz."
          className={history.comebacks.length > 0 ? "lg:row-span-2" : ""}
        >
          <Rows
            limit={10}
            empty="Mindenki versenyzett az elmúlt egy évben."
            rows={history.comebacks.map((comeback) => ({
              key: comeback.person.wcaId ?? comeback.person.name,
              label: <PersonName person={comeback.person} />,
              sub: `Utoljára: ${comeback.lastCompetition.name} (${formatLongDate(comeback.lastCompetition.date)})`,
              value: formatGap(comeback.gapDays),
            }))}
          />
        </StatBlock>

        <StatBlock
          title={`Először versenyez itt: ${country}`}
          note="Visszatérő versenyzők, akiknek ez az első versenyük ebben az országban."
        >
          <Rows
            limit={10}
            empty="Mindenki versenyzett már itt."
            rows={history.firstInCountry.map((person) => ({
              key: person.wcaId ?? person.name,
              label: <PersonName person={person} />,
            }))}
          />
        </StatBlock>

        <StatBlock
          title="Idén a legtöbb versenyen"
          note={`${inYear(Number(competition.startDate.slice(0, 4)))}, ezt a versenyt nem számolva.`}
        >
          <Rows
            empty="Idén ez az első versenyük."
            rows={history.busiestYear.map((entry) => ({
              key: entry.person.wcaId ?? entry.person.name,
              label: <PersonName person={entry.person} />,
              value: `${entry.value} verseny`,
            }))}
          />
        </StatBlock>
      </div>
    </StatSection>
  );
}
