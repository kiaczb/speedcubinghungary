"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  FaArrowLeft,
  FaExclamationTriangle,
  FaExternalLinkAlt,
  FaSpinner,
  FaSyncAlt,
} from "react-icons/fa";
import Badge from "@/app/components/Badge";
import Button from "@/app/components/ui/Button";
import StateMessage from "@/app/components/ui/StateMessage";
import { ApiError, clientJson } from "@/app/lib/api";
import { formatDateRange } from "@/app/lib/date";
import type { Phase, StatsReport } from "../types";
import { countryName } from "../lib/format";
import AgesSection from "./AgesSection";
import CompetitionPicker from "./CompetitionPicker";
import FieldSection from "./FieldSection";
import HistorySection from "./HistorySection";
import LiveSection from "./LiveSection";
import ResultsSection from "./ResultsSection";

const HISTORY_POLL_MS = 4000;
const LIVE_POLL_MS = 60000;

const PHASES: Record<Phase, { label: string; color: string }> = {
  before: {
    label: "Közelgő",
    color: "bg-surface-muted text-muted-foreground ring-line",
  },
  during: {
    label: "Folyamatban",
    color: "bg-brand-red/10 text-brand-red ring-brand-red/20",
  },
  after: {
    label: "Lezajlott",
    color: "bg-brand-green/10 text-brand-green ring-brand-green/20",
  },
};

function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 404) return "Ilyen versenyt nem ismer a WCA.";
    if (error.status === 400) return "Érvénytelen verseny-azonosító.";
    if (error.status === 401 || error.status === 403) return error.message;
  }
  return "A statisztika most nem állítható össze — a WCA nem válaszolt. Próbáld újra kicsit később.";
}

function ReportView({
  competitionId,
  grantFailed,
  onBack,
}: {
  competitionId: string;
  grantFailed: boolean;
  onBack: () => void;
}) {
  const [report, setReport] = useState<StatsReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [refreshes, setRefreshes] = useState(0);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    let cancelled = false;
    clientJson<StatsReport>(
      `/stats/competitions/${encodeURIComponent(competitionId)}`,
    )
      .then((data) => {
        if (cancelled) return;
        setReport(data);
        setError(null);
      })
      .catch((err) => {
        console.error("A statisztika lekérése sikertelen:", err);
        if (!cancelled) setError(errorMessage(err));
      })
      .finally(() => {
        if (!cancelled) setRefreshing(false);
      });
    return () => {
      cancelled = true;
    };
  }, [competitionId, refreshes]);

  const pollDelay =
    report?.history.status === "running"
      ? HISTORY_POLL_MS
      : report?.phase === "during"
        ? LIVE_POLL_MS
        : null;

  useEffect(() => {
    if (!pollDelay) return;
    const timer = setTimeout(() => setRefreshes((n) => n + 1), pollDelay);
    return () => clearTimeout(timer);
  }, [pollDelay, report]);

  const backButton = (
    <Button
      variant="ghost"
      size="sm"
      onClick={onBack}
      icon={<FaArrowLeft className="h-3 w-3" />}
    >
      Másik verseny
    </Button>
  );

  if (!report) {
    return error ? (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="Nem sikerült"
        description={error}
      >
        {backButton}
      </StateMessage>
    ) : (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Statisztika összeállítása…"
        description="Az első betöltés pár másodperc: a nevezőket és az összesítőjüket a WCA-tól kérjük le. Utána gyorsítótárból jön."
      />
    );
  }

  const { competition, phase } = report;
  const phaseBadge = PHASES[phase];
  const sections = {
    results: (
      <ResultsSection
        key="results"
        results={report.results}
        error={report.resultsError}
      />
    ),
    live: (
      <LiveSection
        key="live"
        live={report.live}
        error={report.liveError}
        phase={phase}
      />
    ),
    field: (
      <FieldSection
        key="field"
        competition={competition}
        field={report.field}
        error={report.fieldError}
      />
    ),
    ages: (
      <AgesSection
        key="ages"
        competitionId={competition.id}
        ages={report.ages}
        onChanged={() => setRefreshes((n) => n + 1)}
      />
    ),
    history: (
      <HistorySection
        key="history"
        competition={competition}
        history={report.history}
      />
    ),
  };
  const order: (keyof typeof sections)[] =
    phase === "after"
      ? ["results", "live", "field", "ages", "history"]
      : phase === "during"
        ? ["live", "field", "ages", "history"]
        : ["field", "ages", "history"];

  return (
    <div className="flex flex-col gap-12">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-2xl font-bold text-foreground">
              {competition.name}
            </h2>
            <Badge label={phaseBadge.label} color={phaseBadge.color} />
          </div>
          <p className="tabular mt-1 font-mono text-sm text-muted-foreground">
            {formatDateRange(competition.startDate, competition.endDate)}
            {" · "}
            {competition.city}, {countryName(competition.countryIso2)}
          </p>
          {error && (
            <p className="mt-2 text-sm text-brand-red">
              A frissítés nem sikerült: {error}
            </p>
          )}
          {grantFailed && report.ages.status === "no-grant" && (
            <p className="mt-2 text-sm text-brand-red">
              A WCA-hozzáférés engedélyezése nem sikerült. A saját, admin
              WCA-fiókoddal kell belépni a WCA-n, és el kell fogadni a
              versenykezelési hozzáférést.
            </p>
          )}
        </div>
        <div className="flex flex-wrap gap-2">
          {backButton}
          <Button
            variant="ghost"
            size="sm"
            href={competition.url}
            external
            icon={<FaExternalLinkAlt className="h-3 w-3" />}
          >
            WCA
          </Button>
          <Button
            variant="secondary"
            size="sm"
            disabled={refreshing}
            onClick={() => {
              setRefreshing(true);
              setRefreshes((n) => n + 1);
            }}
            icon={
              <FaSyncAlt
                className={`h-3 w-3 ${refreshing ? "animate-spin" : ""}`}
              />
            }
          >
            Frissítés
          </Button>
        </div>
      </div>

      {order.map((key) => sections[key])}
    </div>
  );
}

export default function StatsView() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const competitionId = searchParams.get("competition");

  if (!competitionId) {
    return (
      <CompetitionPicker
        onSelect={(id) =>
          router.replace(`/admin/stats?competition=${encodeURIComponent(id)}`)
        }
      />
    );
  }

  return (
    <ReportView
      key={competitionId}
      competitionId={competitionId}
      grantFailed={searchParams.get("wca") === "failed"}
      onBack={() => router.replace("/admin/stats")}
    />
  );
}
