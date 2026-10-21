"use client";

import { useRef, useState } from "react";
import { FaExclamationTriangle } from "react-icons/fa";
import "@cubing/icons";
import Card from "@/app/components/Cards/Card";
import { controlClass } from "@/app/components/ui/Field";
import { eventName } from "@/app/lib/events";
import { formatTimeInput, parseTimeInput } from "@/app/lib/time";
import { DNF, DNS, formatAttemptResult, formatCentiseconds } from "../lib/time";
import type { Competitor, CompetitorAttempt, CompetitorRound } from "../types";

type CompetitorCardProps = {
  competitor: Competitor;
  onRecord?: (
    liveResultId: number,
    attemptNumber: number,
    elapsedCs: number | null,
  ) => Promise<void>;
  autoFocus?: boolean;
};

function maxAttemptCount(rounds: CompetitorRound[]): number {
  return rounds.reduce((max, round) => Math.max(max, round.attempts.length), 0);
}

const ReadOnlyAttemptCell = ({ attempt }: { attempt: CompetitorAttempt }) => {
  if (attempt.result === DNF) {
    return (
      <td className="tabular px-2 py-2 text-center font-mono text-sm text-brand-red">
        {attempt.elapsedCs != null ? (
          <span title="DNF · a solve tényleges hossza">
            {formatCentiseconds(attempt.elapsedCs)}
            <span className="ml-1 text-[10px] uppercase">dnf</span>
          </span>
        ) : (
          "DNF"
        )}
      </td>
    );
  }

  return (
    <td
      className={`tabular px-2 py-2 text-center font-mono text-sm ${
        attempt.result === DNS ? "text-subtle-foreground" : "text-foreground"
      }`}
    >
      {formatAttemptResult(attempt.result)}
    </td>
  );
};

type EditableAttemptCellProps = {
  attempt: CompetitorAttempt;
  liveResultId: number;
  onRecord: (
    liveResultId: number,
    attemptNumber: number,
    elapsedCs: number | null,
  ) => Promise<void>;
  autoFocus?: boolean;
};

const EditableAttemptCell = ({
  attempt,
  liveResultId,
  onRecord,
  autoFocus,
}: EditableAttemptCellProps) => {
  const recorded =
    attempt.elapsedCs != null ? formatCentiseconds(attempt.elapsedCs) : "";

  const [value, setValue] = useState(recorded);
  const [busy, setBusy] = useState(false);
  const [invalid, setInvalid] = useState(false);
  const [failed, setFailed] = useState<string | null>(null);

  const savingRef = useRef<string | null>(null);

  const [lastRecorded, setLastRecorded] = useState(recorded);
  if (recorded !== lastRecorded) {
    setLastRecorded(recorded);
    setValue(recorded);
    setFailed(null);
  }

  const save = async () => {
    const trimmed = value.trim();
    const parsed = parseTimeInput(trimmed);

    if (trimmed && parsed === null) {
      setInvalid(true);
      return;
    }

    if (savingRef.current === value) return;
    savingRef.current = value;

    setInvalid(false);
    setFailed(null);
    setBusy(true);
    try {
      await onRecord(liveResultId, attempt.attemptNumber, parsed);
    } catch (error) {
      console.error("A DNF idő mentése sikertelen:", error);
      setFailed(error instanceof Error ? error.message : "Ismeretlen hiba történt.");
    } finally {
      savingRef.current = null;
      setBusy(false);
    }
  };

  return (
    <td className="px-1 py-2 text-center">
      <div className="relative inline-block">
      <input
        value={value}
        onChange={(event) => {
          const formatted = formatTimeInput(event.target.value);
          if (formatted === null) return;

          setValue(formatted);
          setInvalid(false);
        }}
        onBlur={() => {
          if (value !== recorded) void save();
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            void save();
            const inputs = Array.from(
              event.currentTarget
                .closest("table")
                ?.querySelectorAll<HTMLInputElement>("input[data-dnf-input]") ??
                [],
            );
            const next = inputs[inputs.indexOf(event.currentTarget) + 1];
            next?.focus();
            next?.select();
          }
          if (event.key === "Escape") {
            setValue(recorded);
            setInvalid(false);
            setFailed(null);
          }
        }}
        autoFocus={autoFocus}
        data-dnf-input
        disabled={busy}
        placeholder="DNF"
        inputMode="numeric"
        aria-invalid={invalid || failed !== null}
        aria-label={`${attempt.attemptNumber}. kísérlet tényleges ideje`}
        title={
          failed !== null
            ? `NINCS ELMENTVE — ${failed} Próbáld újra (Enter).`
            : attempt.maxElapsedCs != null
              ? `DNF · legfeljebb ${formatCentiseconds(attempt.maxElapsedCs)}`
              : "DNF · nincs időlimit"
        }
        className={`${controlClass} tabular w-22 px-2 py-1.5 text-center font-mono text-sm ${
          invalid || failed !== null
            ? "border-brand-red bg-brand-red/5 text-brand-red"
            : value
              ? "border-brand-green/60 text-foreground"
              : "border-brand-red/50 text-brand-red placeholder:text-brand-red/70"
        }`}
      />

      {failed !== null && (
        <>
          <span
            aria-hidden
            className="pointer-events-none absolute -right-1.5 -top-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-brand-red text-[10px] font-bold leading-none text-white"
          >
            !
          </span>
          <span role="alert" className="sr-only">
            {`A(z) ${attempt.attemptNumber}. kísérlet ideje nincs elmentve: ${failed}`}
          </span>
        </>
      )}
      </div>
    </td>
  );
};

const CompetitorCard = ({
  competitor,
  onRecord,
  autoFocus,
}: CompetitorCardProps) => {
  const columns = maxAttemptCount(competitor.rounds);

  const focusTargetKey = autoFocus
    ? competitor.rounds
        .flatMap((round) =>
          round.attempts
            .filter((attempt) => attempt.result === DNF)
            .map((attempt) => `${round.liveResultId}:${attempt.attemptNumber}`),
        )
        .at(0)
    : undefined;

  return (
    <Card padding="md">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate text-base font-semibold text-foreground">
            {competitor.name}
          </h3>
          <p className="tabular font-mono text-xs text-muted-foreground">
            {competitor.registrantId}
            {competitor.competitorWcaId ? ` · ${competitor.competitorWcaId}` : ""}
          </p>
        </div>

        {competitor.missingDnfCount > 0 && (
          <span className="inline-flex shrink-0 items-center gap-1.5 rounded-full bg-brand-red/10 px-2.5 py-0.5 text-xs font-semibold text-brand-red ring-1 ring-inset ring-brand-red/20">
            <FaExclamationTriangle className="h-3 w-3" aria-hidden />
            {competitor.missingDnfCount} idő hiányzik
          </span>
        )}
      </div>

      <div className="mt-3 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-line text-xs text-subtle-foreground">
              <th className="px-2 py-1.5 text-left font-medium">Kör</th>
              {Array.from({ length: columns }, (_, index) => (
                <th key={index} className="px-2 py-1.5 text-center font-medium">
                  {index + 1}
                </th>
              ))}
              <th className="px-2 py-1.5 text-center font-medium">Átlag</th>
              <th className="px-2 py-1.5 text-center font-medium">Legjobb</th>
            </tr>
          </thead>
          <tbody>
            {competitor.rounds.map((round) => (
              <tr
                key={round.roundWcifId}
                className="border-b border-line last:border-0"
              >
                <td className="tabular whitespace-nowrap px-2 py-2 font-mono text-sm font-medium text-foreground">
                  <span
                    className="inline-flex items-center gap-1.5"
                    title={eventName(round.eventId)}
                  >
                    <span
                      aria-hidden
                      className={`cubing-icon event-${round.eventId} text-lg leading-none text-foreground`}
                    />
                    <span className="sr-only">{round.roundWcifId}</span>
                    <span className="text-xs text-subtle-foreground">
                      {round.roundNumber}
                    </span>
                  </span>
                </td>

                {round.attempts.map((attempt) =>
                  onRecord && attempt.result === DNF ? (
                    <EditableAttemptCell
                      key={attempt.attemptNumber}
                      attempt={attempt}
                      liveResultId={round.liveResultId}
                      onRecord={onRecord}
                      autoFocus={
                        `${round.liveResultId}:${attempt.attemptNumber}` ===
                        focusTargetKey
                      }
                    />
                  ) : (
                    <ReadOnlyAttemptCell
                      key={attempt.attemptNumber}
                      attempt={attempt}
                    />
                  ),
                )}

                {Array.from(
                  { length: columns - round.attempts.length },
                  (_, index) => (
                    <td
                      key={`pad-${index}`}
                      className="px-2 py-2 text-center text-subtle-foreground"
                    >
                      —
                    </td>
                  ),
                )}

                <td className="tabular px-2 py-2 text-center font-mono text-sm text-muted-foreground">
                  {round.average != null ? formatAttemptResult(round.average) : "—"}
                </td>
                <td className="tabular px-2 py-2 text-center font-mono text-sm text-muted-foreground">
                  {round.best != null ? formatAttemptResult(round.best) : "—"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <CompetitorSummary competitor={competitor} />
    </Card>
  );
};

const CompetitorSummary = ({ competitor }: { competitor: Competitor }) => {
  if (competitor.budgets.length === 0) {
    return (
      <div className="mt-3 flex justify-between border-t border-line pt-3 text-sm">
        <span className="text-muted-foreground">Felhasznált</span>
        <span className="tabular font-mono font-semibold text-foreground">
          {formatCentiseconds(competitor.usedCs)}
        </span>
      </div>
    );
  }

  return (
    <div className="mt-3 flex flex-col gap-2 border-t border-line pt-3">
      {competitor.budgets.map((budget) => {
        const exhausted = budget.remainingCs <= 0;

        return (
          <div key={budget.roundWcifIds.join("|")} className="flex flex-col gap-1">
            <div className="flex items-baseline justify-between gap-3 text-sm">
              <span className="text-muted-foreground">
                Maradék
                <span className="tabular ml-1.5 font-mono text-xs text-subtle-foreground">
                  / {formatCentiseconds(budget.limitCs)}
                </span>
              </span>
              <span
                className={`tabular font-mono text-base font-semibold ${
                  exhausted ? "text-brand-red" : "text-brand-green"
                }`}
              >
                {formatCentiseconds(budget.remainingCs)}
              </span>
            </div>

            <div className="flex items-baseline justify-between gap-3 text-xs">
              <span className="text-subtle-foreground">
                Felhasznált{" "}
                <span className="tabular font-mono">
                  {formatCentiseconds(budget.usedCs)}
                </span>
              </span>
              {budget.missingDnfCount > 0 && (
                <span className="text-brand-red">
                  felső becslés · {budget.missingDnfCount} DNF idő nélkül
                </span>
              )}
            </div>

            {budget.roundWcifIds.length > 1 && (
              <p className="tabular font-mono text-[11px] text-subtle-foreground">
                közös keret: {budget.roundWcifIds.join(", ")}
              </p>
            )}
          </div>
        );
      })}
    </div>
  );
};

export default CompetitorCard;
