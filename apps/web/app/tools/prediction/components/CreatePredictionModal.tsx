"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import "@cubing/icons";
import { UpcomingCompetition } from "@/app/calendar/types";
import { clientFetch, clientJson } from "@/app/lib/api";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";
import Modal from "@/app/components/ui/Modal";
import { eventName } from "@/app/lib/events";
import { formatShortDate } from "@/app/lib/date";
import { ExtraQuestion } from "../types";
import ExtraQuestionsEditor from "./ExtraQuestionsEditor";

type Props = {
  onClose: () => void;
};

type CompetitionsStatus = "loading" | "error" | "success";

export default function CreatePredictionModal({ onClose }: Props) {
  const router = useRouter();
  const [competitionsStatus, setCompetitionsStatus] =
    useState<CompetitionsStatus>("loading");
  const [competitions, setCompetitions] = useState<UpcomingCompetition[]>([]);
  const [competitionId, setCompetitionId] = useState("");
  const [podiumEventIds, setPodiumEventIds] = useState<string[]>([]);
  const [predictionDeadline, setPredictionDeadline] = useState("");
  const [extraQuestions, setExtraQuestions] = useState<ExtraQuestion[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const selectRef = useRef<HTMLSelectElement>(null);
  const selectedCompetition = competitions.find((c) => c.id === competitionId);

  useEffect(() => {
    const controller = new AbortController();

    const fetchCompetitions = async () => {
      setCompetitionsStatus("loading");
      try {
        const data = await clientJson<UpcomingCompetition[]>(
          "/competitions/upcoming",
          { signal: controller.signal },
        );
        data.sort((a, b) => a.startDate.localeCompare(b.startDate));

        setCompetitions(data);
        setCompetitionsStatus("success");
        selectRef.current?.focus();
      } catch (err) {
        if (controller.signal.aborted) return;
        console.error("Versenyek lekérése sikertelen:", err);
        setCompetitionsStatus("error");
      }
    };

    fetchCompetitions();
    return () => controller.abort();
  }, []);

  // Verseny váltásakor a dobogó-jelölés a friss esemény-lista szerint marad
  // csak érvényes — egy korábban kiválasztott versenyszám könnyen nem
  // szerepel az új versenyen.
  const handleCompetitionChange = (id: string) => {
    setCompetitionId(id);
    const competition = competitions.find((c) => c.id === id);
    setPodiumEventIds((prev) =>
      prev.filter((eventId) => competition?.eventIds.includes(eventId)),
    );
  };

  const toggleEvent = (eventId: string) => {
    setPodiumEventIds((prev) =>
      prev.includes(eventId)
        ? prev.filter((id) => id !== eventId)
        : [...prev, eventId],
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!competitionId) {
      setError("Válassz versenyt a tippversenyhez.");
      return;
    }
    if (!predictionDeadline) {
      setError("Add meg, meddig lehet tippelni.");
      return;
    }
    if (podiumEventIds.length === 0) {
      setError("Válassz legalább egy versenyszámot a dobogó-tippekhez.");
      return;
    }
    if (extraQuestions.some((q) => !q.prompt.trim())) {
      setError("Minden extra kérdésnek legyen szövege, vagy töröld a mezőt.");
      return;
    }

    setSubmitting(true);
    try {
      await clientFetch("/predictions", {
        method: "POST",
        json: {
          competitionId,
          predictionDeadline: new Date(predictionDeadline).toISOString(),
          podiumEventIds,
          extraQuestions: extraQuestions.map((q) => ({
            prompt: q.prompt.trim(),
            type: q.type,
          })),
        },
      });

      router.refresh();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title="Új tippverseny létrehozása"
      onClose={onClose}
      size="lg"
      closeDisabled={submitting}
      initialFocusRef={selectRef}
    >
      <form onSubmit={handleSubmit} className="space-y-5 px-6 py-5">
        <Field label="Verseny" htmlFor="prediction-competition">
          {competitionsStatus === "error" ? (
            <p className="rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong">
              Nem sikerült betölteni a versenyeket.
            </p>
          ) : (
            <select
              id="prediction-competition"
              ref={selectRef}
              value={competitionId}
              onChange={(e) => handleCompetitionChange(e.target.value)}
              disabled={competitionsStatus === "loading"}
              className={controlClass}
            >
              <option value="">
                {competitionsStatus === "loading"
                  ? "Versenyek betöltése…"
                  : "Válassz versenyt"}
              </option>
              {competitions.map((competition) => (
                <option key={competition.id} value={competition.id}>
                  {competition.name} —{" "}
                  {formatShortDate(competition.startDate)}
                </option>
              ))}
            </select>
          )}
        </Field>

        <Field
          label="Tippelési határidő"
          htmlFor="prediction-deadline"
          hint="Eddig lehet tippet leadni — utána a tippverseny lezárul."
        >
          <input
            id="prediction-deadline"
            type="datetime-local"
            value={predictionDeadline}
            onChange={(e) => setPredictionDeadline(e.target.value)}
            className={controlClass}
          />
        </Field>

        <Field
          label="Dobogó-tippek versenyszámai"
          htmlFor="prediction-events"
          hint={
            selectedCompetition
              ? "Ezekhez a versenyszámokhoz kell majd 1-2-3. helyet tippelni."
              : "Válassz versenyt — a lista a verseny versenyszámaiból töltődik."
          }
        >
          {selectedCompetition ? (
            <div
              id="prediction-events"
              className="flex flex-wrap gap-2"
              role="group"
              aria-label="Dobogó-tippek versenyszámai"
            >
              {selectedCompetition.eventIds.map((eventId) => {
                const checked = podiumEventIds.includes(eventId);
                return (
                  <label
                    key={eventId}
                    className={`flex cursor-pointer items-center gap-2 rounded-control border px-3 py-2 text-sm transition-colors duration-200 ease-smooth ${
                      checked
                        ? "border-foreground bg-surface-muted text-foreground"
                        : "border-line text-muted-foreground hover:border-line-strong"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={() => toggleEvent(eventId)}
                      className="h-3.5 w-3.5 accent-foreground"
                    />
                    <span
                      className={`cubing-icon event-${eventId} text-[15px] leading-none`}
                      aria-hidden
                    />
                    {eventName(eventId)}
                  </label>
                );
              })}
            </div>
          ) : (
            <p className="text-sm text-subtle-foreground">
              Nincs kiválasztott verseny.
            </p>
          )}
        </Field>

        <Field
          label="Extra kérdések"
          htmlFor="prediction-extra-questions"
          optional
          hint="Pl. versenyző, idő, szám vagy szabad szöveg tippelése — például 'Ki nyeri a 3x3x3-at?' vagy 'Hány rekord születik?'."
        >
          <ExtraQuestionsEditor
            questions={extraQuestions}
            onChange={setExtraQuestions}
          />
        </Field>

        {error && (
          <p
            role="alert"
            className="rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong"
          >
            {error}
          </p>
        )}

        <div className="flex justify-end gap-3 border-t border-line pt-5">
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={submitting}
          >
            Mégse
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Létrehozás…" : "Létrehozás"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
