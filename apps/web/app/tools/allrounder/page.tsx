import type { Metadata } from "next";
import { FaExclamationTriangle, FaTrophy } from "react-icons/fa";
import PageHeader from "@/app/components/ui/PageHeader";
import StateMessage from "@/app/components/ui/StateMessage";
import { serverJson } from "@/app/lib/api";
import { formatTimestamp } from "@/app/lib/date";
import LiveBanner from "./components/LiveBanner";
import RefreshButton from "./components/RefreshButton";
import RulesModal from "./components/RulesModal";
import StandingsList from "./components/StandingsList";
import StandingsTable from "./components/StandingsTable";
import YearPicker from "./components/YearPicker";
import type { Standings } from "./types";

export const metadata: Metadata = {
  title: "Hungarian All-Rounder",
  description:
    "A Hungarian All-Rounder versenysorozat pontversenyének állása évadonként, a WCA hivatalos eredményei alapján.",
};

type AllRounderPageProps = {
  searchParams: Promise<{ year?: string | string[] }>;
};

async function getStandings(year?: string): Promise<Standings | null> {
  const query = year ? `?year=${encodeURIComponent(year)}` : "";

  return serverJson<Standings | null>(`/allrounder/standings${query}`, null);
}

export default async function AllRounderPage({
  searchParams,
}: AllRounderPageProps) {
  const { year } = await searchParams;
  const standings = await getStandings(
    typeof year === "string" ? year : undefined,
  );

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-7xl">
        <PageHeader
          title="Hungarian All-Rounder"
          description="A Hungarian All-Rounder egy állomásokból álló versenysorozat: minden állomás önálló WCA-verseny, a sorozat végső rangsorát pedig az állomásokon szerzett pontok összege adja. Az itt látható pontverseny a WCA hivatalos eredményeiből számolódik."
        />

        {!standings ? (
          <div className="mt-10">
            <StateMessage
              icon={<FaExclamationTriangle className="h-5 w-5" />}
              title="Nem sikerült betölteni"
              description="A pontverseny állása most nem érhető el. Próbáld újra kicsit később."
            />
          </div>
        ) : (
          <>
            <div className="mt-8 flex flex-wrap items-center justify-between gap-4">
              <YearPicker
                key={standings.year}
                selectedYear={standings.year}
                availableYears={standings.availableYears}
              />

              <div className="flex items-center gap-2">
                <RefreshButton year={standings.year} />
                <RulesModal
                  pointsTable={standings.pointsTable}
                  countingResults={standings.countingResults}
                  podiumCount={standings.podiumCount}
                />
              </div>
            </div>

            <LiveBanner year={standings.year} live={standings.live} />

            <div className="mt-6">
              {standings.people.length === 0 ? (
                <StateMessage
                  icon={<FaTrophy className="h-5 w-5" />}
                  title="Még nincs eredmény"
                  description="Ennek az évadnak még egyetlen állomásáról sem érkezett hivatalos WCA-eredmény."
                />
              ) : (
                <>
                  <div className="md:hidden">
                    <StandingsList
                      competitions={standings.competitions}
                      people={standings.people}
                      podiumCount={standings.podiumCount}
                    />
                  </div>

                  <div className="hidden md:block">
                    <StandingsTable
                      competitions={standings.competitions}
                      people={standings.people}
                      podiumCount={standings.podiumCount}
                    />
                  </div>
                </>
              )}
            </div>

            <div className="mt-4 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-2 text-xs leading-relaxed text-muted-foreground">
              <p className="max-w-3xl">
                Minden versenyzőnél az állomásonként szerzett, beszámító
                eredményei állnak: a versenyszám ikonja és a helyezése.
                Telefonon a versenyző sorára koppintva nyílik ki az
                állomásonkénti bontás. A sorozat végén az első{" "}
                {standings.podiumCount} helyezettet díjazzuk — őket a sor bal
                szélén futó vonal jelöli.
              </p>

              {standings.updatedAt && (
                <p
                  className="tabular shrink-0 font-mono"
                  title="A sorozat állomásai közül a legrégebbi lekérés ideje — ennél régebbi adat nincs a táblázatban."
                >
                  Eredmények frissítve:{" "}
                  {formatTimestamp(standings.updatedAt)}
                </p>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
