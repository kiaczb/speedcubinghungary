"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  FaCircle,
  FaExclamationTriangle,
  FaExchangeAlt,
  FaSatelliteDish,
  FaSearch,
  FaSpinner,
} from "react-icons/fa";
import { useUser } from "@/app/components/UserProvider";
import Button from "@/app/components/ui/Button";
import { controlClass } from "@/app/components/ui/Field";
import StateMessage from "@/app/components/ui/StateMessage";
import { clientApiUrl, clientFetch, clientJson } from "@/app/lib/api";
import CompetitionPicker from "./CompetitionPicker";
import CompetitorCard from "./CompetitorCard";
import RoundTabs from "./RoundTabs";
import type { Competition, Competitor, SpecialReport } from "../types";

const FALLBACK_POLL_MS = 30_000;

function fetchReport(
  competition: string,
  round: number | null,
): Promise<SpecialReport> {
  const roundParam = round === null ? "" : `&round=${round}`;
  return clientJson<SpecialReport>(
    `/special/report?competition=${encodeURIComponent(competition)}${roundParam}`,
  );
}

function matchesQuery(competitor: Competitor, query: string): boolean {
  const needle = query.trim().toLowerCase();
  if (!needle) return true;

  return (
    competitor.name.toLowerCase().includes(needle) ||
    String(competitor.registrantId).includes(needle) ||
    (competitor.competitorWcaId ?? "").toLowerCase().includes(needle)
  );
}

const SpecialView = () => {
  const { isAdmin } = useUser();

  const router = useRouter();
  const searchParams = useSearchParams();
  const competitionId = searchParams.get("competition");

  const [report, setReport] = useState<SpecialReport | null>(null);
  const [failedCompetition, setFailedCompetition] = useState<string | null>(
    null,
  );
  const [live, setLive] = useState(false);
  const [query, setQuery] = useState("");
  const [onlyMissing, setOnlyMissing] = useState(false);
  const [roundSelection, setRoundSelection] = useState<{
    competition: string | null;
    round: number | null;
  }>({ competition: null, round: null });

  const selectedRound =
    roundSelection.competition === competitionId ? roundSelection.round : null;

  const setSelectedRound = (round: number | null) => {
    setRoundSelection({ competition: competitionId, round });
  };

  const selectedRoundRef = useRef<number | null>(null);

  const selectCompetition = (competition: Competition) => {
    router.replace(
      `/tools/special?competition=${encodeURIComponent(competition.competitionWcaId)}`,
      { scroll: false },
    );
  };

  const clearCompetition = () => {
    router.replace("/tools/special", { scroll: false });
  };

  useEffect(() => {
    selectedRoundRef.current = selectedRound;

    if (!competitionId) return;

    let cancelled = false;

    fetchReport(competitionId, selectedRound)
      .then((data) => {
        if (cancelled) return;
        setReport(data);
        setFailedCompetition(null);
      })
      .catch((error) => {
        console.error("A Special riport lekérése sikertelen:", error);
        if (!cancelled) setFailedCompetition(competitionId);
      });

    return () => {
      cancelled = true;
    };
  }, [competitionId, selectedRound]);

  useEffect(() => {
    if (!competitionId) return;

    const onUpdate = () => {
      fetchReport(competitionId, selectedRoundRef.current)
        .then((data) => {
          setReport(data);
          setFailedCompetition(null);
        })
        .catch((error) => {
          console.error("A Special riport lekérése sikertelen:", error);
        });
    };

    const source = new EventSource(
      `${clientApiUrl()}/special/stream?competition=${encodeURIComponent(competitionId)}`,
      { withCredentials: true },
    );

    source.addEventListener("ready", () => setLive(true));
    source.addEventListener("update", onUpdate);
    source.onerror = () => setLive(false);

    const fallback = setInterval(onUpdate, FALLBACK_POLL_MS);

    return () => {
      source.close();
      clearInterval(fallback);
      setLive(false);
    };
  }, [competitionId]);

  const handleRecord = async (
    liveResultId: number,
    attemptNumber: number,
    elapsedCs: number | null,
  ) => {
    if (!competitionId) return;

    // A mentés hibája SZÁNDÉKOSAN felszáll a hívó cellához: az jelzi a
    // rögzítőnek, hogy a beírt idő nincs elmentve (lásd CompetitorCard).
    await clientFetch(
      `/special/results/${liveResultId}/attempts/${attemptNumber}/elapsed`,
      {
        method: "PUT",
        json: { competitionWcaId: competitionId, elapsedCs },
      },
    );

    // A visszaolvasás hibája viszont NEM mentési hiba — a mentés ekkor már
    // megtörtént. Ha ezt is felengednénk, a cella mentetlennek jelölné a már
    // elmentett időt. Az SSE (vagy a tartalék poll) úgyis frissít.
    try {
      setReport(await fetchReport(competitionId, selectedRound));
    } catch (error) {
      console.error("A Special riport frissítése a mentés után sikertelen:", error);
    }
  };

  if (!competitionId) {
    return <CompetitionPicker onSelect={selectCompetition} />;
  }

  const current =
    report?.competition.competitionWcaId === competitionId ? report : null;

  if (!current && failedCompetition === competitionId) {
    return (
      <div className="flex flex-col items-center gap-4">
        <StateMessage
          icon={<FaExclamationTriangle className="h-5 w-5" />}
          title="Nem sikerült betölteni"
          description="Lehet, hogy nincs ilyen verseny a WCA Live-on, vagy a szolgáltatás épp nem elérhető."
        />
        <Button variant="secondary" size="sm" onClick={clearCompetition}>
          Másik verseny választása
        </Button>
      </div>
    );
  }

  if (!current) {
    return (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Betöltés…"
        description="Eredmények lekérése a WCA Live-ról."
      />
    );
  }

  const competition = current.competition;
  const competitors = current.competitors;
  const availableRounds = current.availableRounds;

  const totalMissing = competitors.reduce(
    (sum, competitor) => sum + competitor.missingDnfCount,
    0,
  );

  const missingFilterAvailable = isAdmin && totalMissing > 0;

  const visible = competitors
    .filter((competitor) => matchesQuery(competitor, query))
    .filter(
      (competitor) =>
        !(missingFilterAvailable && onlyMissing) ||
        competitor.missingDnfCount > 0,
    );

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-baseline justify-between gap-3 border-b border-line pb-4">
        <div className="min-w-0">
          <h2 className="text-xl font-bold text-foreground">
            {competition.name}
          </h2>
          <p className="tabular font-mono text-xs text-muted-foreground">
            {competition.startDate}
          </p>
        </div>

        <div className="flex items-center gap-4">
          <span
            className="flex items-center gap-1.5 text-xs text-muted-foreground"
            title={
              live
                ? "Élő kapcsolat — az eredmények maguktól frissülnek."
                : "Nincs élő kapcsolat; időzített frissítés van érvényben."
            }
          >
            <FaCircle
              className={`h-2 w-2 ${live ? "text-brand-green" : "text-subtle-foreground"}`}
              aria-hidden
            />
            {live ? "Élő" : "Offline"}
          </span>

          <Button
            variant="ghost"
            size="sm"
            onClick={clearCompetition}
            icon={<FaExchangeAlt className="h-3 w-3" />}
          >
            Másik verseny
          </Button>
        </div>
      </div>

      <RoundTabs
        rounds={availableRounds}
        selected={selectedRound}
        onSelect={setSelectedRound}
      />

      {competitors.length === 0 ? (
        <StateMessage
          icon={<FaSatelliteDish className="h-5 w-5" />}
          title={
            selectedRound === null
              ? "Még nincs eredmény"
              : `A(z) ${selectedRound}. körben még nincs eredmény`
          }
          description={
            selectedRound === null
              ? "Amint a versenyen eredményt rögzítenek, itt megjelenik — magától, frissítés nélkül."
              : "Válassz másik kört, vagy várd meg, amíg ebben a körben is rögzítenek eredményt."
          }
        />
      ) : (
        <div className="flex flex-col gap-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="relative flex-1 sm:max-w-sm">
              <FaSearch
                className="pointer-events-none absolute left-3.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-subtle-foreground"
                aria-hidden
              />
              <input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Név vagy rajtszám…"
                aria-label="Versenyző keresése"
                className={`${controlClass} pl-9`}
              />
            </div>

            {missingFilterAvailable && (
              <label className="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground">
                <input
                  type="checkbox"
                  checked={onlyMissing}
                  onChange={(event) => setOnlyMissing(event.target.checked)}
                  className="h-4 w-4 rounded border-line-strong"
                />
                Csak hiányzó idők
                <span className="tabular font-mono font-semibold text-brand-red">
                  ({totalMissing})
                </span>
              </label>
            )}
          </div>

          {visible.length === 0 ? (
            <StateMessage
              icon={<FaSearch className="h-5 w-5" />}
              title="Nincs találat"
              description="Módosítsd a keresést vagy kapcsold ki a szűrőt."
            />
          ) : (
            <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
              {visible.map((competitor) => (
                <CompetitorCard
                  key={competitor.registrantId}
                  competitor={competitor}
                  onRecord={isAdmin ? handleRecord : undefined}
                  autoFocus={
                    isAdmin && visible.length === 1 && query.trim().length > 0
                  }
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default SpecialView;
