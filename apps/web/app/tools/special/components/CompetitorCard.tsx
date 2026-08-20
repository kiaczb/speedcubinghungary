"use client";

import { useState } from "react";
import { FaExclamationTriangle } from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import { controlClass } from "@/app/components/ui/Field";
import {
  formatAttemptResult,
  formatCentiseconds,
  formatTimeInput,
  parseTimeToCentiseconds,
} from "../lib/time";
import type { Competitor, CompetitorAttempt, CompetitorRound } from "../types";

// WCIF AttemptResult speciális értékei.
const DNF = -1;
const DNS = -2;

type CompetitorCardProps = {
  competitor: Competitor;
  /**
   * Csak adminnak: az idő rögzítése. Enélkül a kártya csak olvasható.
   * A kulcs a WCA Live result id-ja + a kísérlet sorszáma — a DNF-ek listája a
   * friss WCA Live adatból jön, nálunk csak a rögzített idő él.
   */
  onRecord?: (
    liveResultId: number,
    attemptNumber: number,
    elapsedCs: number | null,
  ) => Promise<void>;
  /** A kártya első DNF-mezőjére fókuszál (kereséskor). */
  autoFocus?: boolean;
};

// A fejléc a leghosszabb kör kísérletszámához igazodik (Bo3/Mo3-nál kevesebb).
function maxAttemptCount(rounds: CompetitorRound[]): number {
  return rounds.reduce((max, round) => Math.max(max, round.attempts.length), 0);
}

/** Nem szerkeszthető cella: hivatalos eredmény, vagy DNF a rögzített idővel. */
const ReadOnlyAttemptCell = ({ attempt }: { attempt: CompetitorAttempt }) => {
  if (attempt.result === DNF) {
    return (
      <td className="tabular px-2 py-2 text-center font-mono text-sm text-brand-red">
        {attempt.elapsedCs != null ? (
          // A rögzített idő a MI megfigyelésünk, nem hivatalos eredmény —
          // ezért jelöljük, hogy DNF-ről van szó.
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

  // Ha máshonnan (másik gépről, vagy a háttérfrissítésből) változik a rögzített
  // érték, átvesszük — renderben igazítva, nem effectben.
  const [lastRecorded, setLastRecorded] = useState(recorded);
  if (recorded !== lastRecorded) {
    setLastRecorded(recorded);
    setValue(recorded);
  }

  const save = async () => {
    const trimmed = value.trim();
    const parsed = parseTimeToCentiseconds(trimmed);

    if (trimmed && parsed === null) {
      setInvalid(true);
      return;
    }

    setInvalid(false);
    setBusy(true);
    try {
      await onRecord(liveResultId, attempt.attemptNumber, parsed);
    } finally {
      setBusy(false);
    }
  };

  return (
    <td className="px-1 py-2 text-center">
      <input
        value={value}
        onChange={(event) => {
          setValue(formatTimeInput(event.target.value));
          setInvalid(false);
        }}
        onBlur={() => {
          if (value !== recorded) void save();
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            void save();
            // Enter után a következő mezőre lépünk, ahogy a versenyeken
            // használt beviteli felületeken szokás.
            // Kártyánként egy táblázat van, ezért ez pontosan az adott
            // versenyző mezőit adja vissza.
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
          }
        }}
        autoFocus={autoFocus}
        data-dnf-input
        disabled={busy}
        placeholder="DNF"
        inputMode="numeric"
        aria-label={`${attempt.attemptNumber}. kísérlet tényleges ideje`}
        title={
          attempt.maxElapsedCs != null
            ? `DNF · legfeljebb ${formatCentiseconds(attempt.maxElapsedCs)}`
            : "DNF · nincs időlimit"
        }
        className={`${controlClass} tabular w-22 px-2 py-1.5 text-center font-mono text-sm ${
          invalid
            ? "border-brand-red text-brand-red"
            : value
              ? "border-brand-green/60 text-foreground"
              : "border-brand-red/50 text-brand-red placeholder:text-brand-red/70"
        }`}
      />
    </td>
  );
};

const CompetitorCard = ({
  competitor,
  onRecord,
  autoFocus,
}: CompetitorCardProps) => {
  const columns = maxAttemptCount(competitor.rounds);

  // Kereséskor az első DNF-mezőre ugrunk. A kísérletet a kör result id-ja és a
  // sorszám azonosítja, ezért ebből képezünk kulcsot.
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
                  {round.roundWcifId}
                </td>

                {round.attempts.map((attempt) =>
                  // Csak a DNF-ből lesz beviteli mező; minden más cella
                  // állapot nélküli megjelenítés.
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

// Összesítő: a BpSpecial "Remaining / Used" sávjának megfelelője. Kumulatív
// keret nélkül csak a felhasznált időt tudjuk mutatni — maradék nincs.
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
