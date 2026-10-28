"use client";

import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import {
  FaArrowLeft,
  FaExclamationTriangle,
  FaExternalLinkAlt,
  FaHourglassHalf,
  FaSpinner,
  FaSyncAlt,
  FaUsers,
} from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import StateMessage from "@/app/components/ui/StateMessage";
import { ApiError, clientJson } from "@/app/lib/api";
import { formatDateRange, formatDayMonthTime, formatLongDate } from "@/app/lib/date";
import { countryName } from "../../stats/lib/format";
import type { CertificateMode, Recipient, RecipientList } from "../types";
import { modesFor, recipientKey, sortByName } from "../lib/recipients";
import type { CertificateTemplate } from "../lib/template/model";
import GenerateBar from "./GenerateBar";
import ModeTabs from "./ModeTabs";
import PeopleList from "./PeopleList";
import PodiumList from "./PodiumList";

type Loaded = {
  mode: CertificateMode;
  list: RecipientList;
};

type Failure = {
  mode: CertificateMode;
  message: string;
};

function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 404) return "Ilyen versenyt nem ismer a WCA.";
    if (error.status === 400) return "Érvénytelen verseny-azonosító.";
    if (error.status === 401 || error.status === 403) return error.message;
  }
  return "A WCA vagy a WCA Live most nem válaszolt, vagy a verseny nincs fent a WCA Live-on. Próbáld újra kicsit később.";
}

function sourceLabel(list: RecipientList): string {
  switch (list.source) {
    case "results":
      return "Forrás: hivatalos WCA-eredmények";
    case "live":
      return list.fetchedAt
        ? `Forrás: WCA Live, lekérve ${formatDayMonthTime(list.fetchedAt)}`
        : "Forrás: WCA Live";
    case "registrations":
      return "Forrás: WCA nevezési lista";
    default:
      return "";
  }
}

export default function CompetitionCertificates({
  template,
  competitionId,
  mode,
  onModeChange,
  onBack,
}: {
  template: CertificateTemplate;
  competitionId: string;
  mode: CertificateMode;
  onModeChange: (mode: CertificateMode) => void;
  onBack: () => void;
}) {
  const [loaded, setLoaded] = useState<Loaded | null>(null);
  const [failure, setFailure] = useState<Failure | null>(null);
  const [refreshes, setRefreshes] = useState(0);
  const [refreshing, setRefreshing] = useState(false);
  const [excluded, setExcluded] = useState<Set<string>>(() => new Set());
  const [excludedEvents, setExcludedEvents] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    let cancelled = false;
    const params = new URLSearchParams({ competition: competitionId, mode });
    clientJson<RecipientList>(`/certificates/recipients?${params}`)
      .then((list) => {
        if (cancelled) return;
        setLoaded({ mode, list });
        setFailure(null);
      })
      .catch((error) => {
        console.error("Az oklevél-címzettek lekérése sikertelen:", error);
        if (!cancelled) setFailure({ mode, message: errorMessage(error) });
      })
      .finally(() => {
        if (!cancelled) setRefreshing(false);
      });
    return () => {
      cancelled = true;
    };
  }, [competitionId, mode, refreshes]);

  const toggle = (set: Set<string>, key: string) => {
    const next = new Set(set);
    if (!next.delete(key)) next.add(key);
    return next;
  };
  const excludedKey = (recipient: Recipient) => `${mode}|${recipientKey(recipient)}`;
  const isExcluded = (recipient: Recipient) => excluded.has(excludedKey(recipient));
  const onToggle = (recipient: Recipient) =>
    setExcluded((current) => toggle(current, excludedKey(recipient)));

  const modes = modesFor(template.kind);
  const canRegistrants = modes.some((m) => m.mode === "registrants");
  const competition = loaded?.list.competition;
  const current = loaded?.mode === mode ? loaded.list : null;
  const currentFailure = failure?.mode === mode ? failure.message : null;

  const recipients = current
    ? mode === "podium"
      ? current.recipients
      : sortByName(current.recipients)
    : [];
  const selected = recipients.filter(
    (r) => !isExcluded(r) && !(r.eventId && excludedEvents.has(r.eventId)),
  );

  const backButton = (
    <Button variant="ghost" size="sm" onClick={onBack} icon={<FaArrowLeft className="h-3 w-3" />}>
      Másik verseny
    </Button>
  );

  let body: ReactNode;
  if (!current && currentFailure) {
    body = (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="Nem sikerült lekérni a listát"
        description={currentFailure}
      >
        {backButton}
      </StateMessage>
    );
  } else if (!current) {
    body = (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Lista összeállítása…"
        description="A WCA-tól, versenyközben a WCA Live-ról kérjük le."
      />
    );
  } else if (current.source === "none") {
    body = (
      <StateMessage
        icon={<FaHourglassHalf className="h-5 w-5" />}
        title="Még nincs eredmény"
        description={`A verseny még nem kezdődött el (kezdés: ${formatLongDate(current.competition.startDate)}), ezért ${canRegistrants ? "versenyzők még nincsenek. A regisztráltak listája már elérhető." : "dobogósok még nincsenek."}`}
      >
        {canRegistrants ? (
          <Button variant="secondary" size="sm" onClick={() => onModeChange("registrants")}>
            Regisztráltak listája
          </Button>
        ) : (
          backButton
        )}
      </StateMessage>
    );
  } else if (mode === "podium") {
    body = (
      <PodiumList
        events={current.events}
        recipients={recipients}
        isExcluded={isExcluded}
        onToggle={onToggle}
        isEventExcluded={(eventId) => excludedEvents.has(eventId)}
        onToggleEvent={(eventId) => setExcludedEvents((current) => toggle(current, eventId))}
      />
    );
  } else if (recipients.length === 0) {
    body = (
      <StateMessage
        icon={<FaUsers className="h-5 w-5" />}
        title="Üres a lista"
        description="Ehhez a versenyhez ebben a módban nincs egyetlen név sem."
      />
    );
  } else {
    body = <PeopleList recipients={recipients} isExcluded={isExcluded} onToggle={onToggle} />;
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">
            {competition?.name ?? competitionId}
          </h2>
          {competition && (
            <p className="tabular mt-1 font-mono text-sm text-muted-foreground">
              {formatDateRange(competition.startDate, competition.endDate)}
              {" · "}
              {competition.city}, {countryName(competition.countryIso2)}
            </p>
          )}
        </div>
        <div className="flex flex-wrap gap-2">
          {backButton}
          {competition && (
            <Button
              variant="ghost"
              size="sm"
              href={competition.url}
              external
              icon={<FaExternalLinkAlt className="h-3 w-3" />}
            >
              WCA
            </Button>
          )}
          <Button
            variant="secondary"
            size="sm"
            disabled={refreshing || !current}
            onClick={() => {
              setRefreshing(true);
              setRefreshes((n) => n + 1);
            }}
            icon={<FaSyncAlt className={`h-3 w-3 ${refreshing ? "animate-spin" : ""}`} />}
          >
            Frissítés
          </Button>
        </div>
      </div>

      <ModeTabs modes={modes} mode={mode} onChange={onModeChange} />

      {current && current.source !== "none" && (
        <div className="-mt-4 flex flex-col gap-1">
          <p className="text-xs text-subtle-foreground">{sourceLabel(current)}</p>
          {currentFailure && (
            <p className="text-sm text-brand-red">A frissítés nem sikerült: {currentFailure}</p>
          )}
        </div>
      )}

      {body}

      {current && competition && recipients.length > 0 && (
        <GenerateBar
          key={mode}
          template={template}
          competition={competition}
          mode={mode}
          selected={selected}
          total={recipients.length}
        />
      )}
    </div>
  );
}
