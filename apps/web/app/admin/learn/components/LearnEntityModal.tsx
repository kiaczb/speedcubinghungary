"use client";
import { useEffect, useRef, useState } from "react";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";
import type { LearnEventNode } from "@/app/learn/types";
import {
  createAlgorithm,
  createCase,
  createEvent,
  updateAlgorithm,
  updateCase,
  updateEvent,
} from "../lib/api";
import type { LearnModalState } from "../types";

// Egyezzenek a backend korlátaival: apps/api/internal/learn/validator.go
const MAX_NAME_LENGTH = 120;
const MAX_DESCRIPTION_LENGTH = 2000;
const MAX_NOTATION_LENGTH = 500;
const MAX_VIDEO_URL_LENGTH = 500;

type Props = {
  state: LearnModalState;
  tree: LearnEventNode[];
  onClose: () => void;
  /** A szülő ebből tölti újra a fát és mutatja a visszajelzést. */
  onSaved: (message: string) => void;
};

const TITLES: Record<LearnModalState["kind"], { create: string; edit: string }> =
  {
    event: { create: "Új versenyszám", edit: "Versenyszám szerkesztése" },
    case: { create: "Új eset", edit: "Eset szerkesztése" },
    algorithm: { create: "Új algoritmus", edit: "Algoritmus szerkesztése" },
  };

export default function LearnEntityModal({
  state,
  tree,
  onClose,
  onSaved,
}: Props) {
  const isEdit = state.entity !== null;

  const [name, setName] = useState(state.entity?.name ?? "");
  const [description, setDescription] = useState(
    state.entity?.description ?? "",
  );
  // Sorrend szövegként: az üres mező így nem ugrik magától 0-ra gépelés közben.
  const [order, setOrder] = useState(String(state.entity?.order ?? 0));
  const [notation, setNotation] = useState(
    state.kind === "algorithm" ? (state.entity?.notation ?? "") : "",
  );
  const [videoUrl, setVideoUrl] = useState(
    state.kind === "algorithm" ? (state.entity?.videoUrl ?? "") : "",
  );
  const [eventId, setEventId] = useState(
    state.kind === "case" ? state.eventId : "",
  );
  const [caseId, setCaseId] = useState(
    state.kind === "algorithm" ? state.caseId : "",
  );

  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const nameInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  useEffect(() => {
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    nameInputRef.current?.focus();
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, []);

  // Az algoritmus eset-választójához a teljes lista kell, versenyszámmal együtt
  // — így egyértelmű, melyik "OLL 21"-ről van szó.
  const allCases = tree.flatMap((event) =>
    event.cases.map((item) => ({
      id: item.id,
      label: `${event.name} → ${item.name}`,
    })),
  );

  const validate = (): string | null => {
    if (!name.trim()) return "A név nem lehet üres.";
    if (name.trim().length > MAX_NAME_LENGTH)
      return `A név legfeljebb ${MAX_NAME_LENGTH} karakter lehet.`;
    if (description.length > MAX_DESCRIPTION_LENGTH)
      return `A leírás legfeljebb ${MAX_DESCRIPTION_LENGTH} karakter lehet.`;
    if (!Number.isInteger(Number(order)))
      return "A sorrend egész szám legyen.";

    if (state.kind === "case" && !eventId)
      return "Válaszd ki, melyik versenyszámhoz tartozik.";

    if (state.kind === "algorithm") {
      if (!caseId) return "Válaszd ki, melyik esethez tartozik.";
      if (!notation.trim()) return "Az algoritmus nem lehet üres.";
      if (notation.trim().length > MAX_NOTATION_LENGTH)
        return `Az algoritmus legfeljebb ${MAX_NOTATION_LENGTH} karakter lehet.`;

      const url = videoUrl.trim();
      if (url) {
        if (url.length > MAX_VIDEO_URL_LENGTH)
          return `A videó URL legfeljebb ${MAX_VIDEO_URL_LENGTH} karakter lehet.`;
        // Ugyanaz a szabály, amit a backend is kikényszerít: csak abszolút
        // http(s) URL — így nem kerülhet `javascript:` a lejátszó helyére.
        try {
          const parsed = new URL(url);
          if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
            return "A videó URL csak http vagy https címre mutathat.";
          }
        } catch {
          return "A videó URL érvénytelen. Teljes címet adj meg (https://…).";
        }
      }
    }

    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }

    const shared = {
      name: name.trim(),
      description: description.trim(),
      order: Number(order),
    };

    setSubmitting(true);
    setError(null);

    try {
      if (state.kind === "event") {
        if (state.entity) await updateEvent(state.entity.id, shared);
        else await createEvent(shared);
      } else if (state.kind === "case") {
        const payload = { ...shared, eventId };
        if (state.entity) await updateCase(state.entity.id, payload);
        else await createCase(payload);
      } else {
        const payload = {
          ...shared,
          caseId,
          notation: notation.trim(),
          videoUrl: videoUrl.trim(),
        };
        if (state.entity) await updateAlgorithm(state.entity.id, payload);
        else await createAlgorithm(payload);
      }

      onSaved(isEdit ? "A módosítás elmentve." : "Sikeresen létrehozva.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
    } finally {
      setSubmitting(false);
    }
  };

  const titles = TITLES[state.kind];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm dark:bg-black/70"
      onClick={onClose}
      role="presentation"
    >
      <div
        className="max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-panel border border-line bg-surface shadow-panel"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="learn-modal-title"
      >
        <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-line bg-surface/95 px-6 py-4 backdrop-blur-sm">
          <h2 id="learn-modal-title" className="text-lg font-bold">
            {isEdit ? titles.edit : titles.create}
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Bezárás"
            className="-mr-2 -mt-1 inline-flex h-8 w-8 items-center justify-center rounded-control text-xl leading-none text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground"
          >
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5 px-6 py-5">
          {state.kind === "case" && (
            <Field label="Versenyszám" htmlFor="learn-event">
              <select
                id="learn-event"
                value={eventId}
                onChange={(e) => setEventId(e.target.value)}
                className={controlClass}
              >
                <option value="">Válassz versenyszámot…</option>
                {tree.map((event) => (
                  <option key={event.id} value={event.id}>
                    {event.name}
                  </option>
                ))}
              </select>
            </Field>
          )}

          {state.kind === "algorithm" && (
            <Field label="Eset" htmlFor="learn-case">
              <select
                id="learn-case"
                value={caseId}
                onChange={(e) => setCaseId(e.target.value)}
                className={controlClass}
              >
                <option value="">Válassz esetet…</option>
                {allCases.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.label}
                  </option>
                ))}
              </select>
            </Field>
          )}

          <Field label="Név" htmlFor="learn-name">
            <input
              id="learn-name"
              ref={nameInputRef}
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={MAX_NAME_LENGTH}
              placeholder={
                state.kind === "event"
                  ? "pl. 3x3x3"
                  : state.kind === "case"
                    ? "pl. OLL 21"
                    : "pl. Sune"
              }
              className={controlClass}
            />
          </Field>

          {state.kind === "algorithm" && (
            <Field
              label="Algoritmus"
              htmlFor="learn-notation"
              hint="A lépéssorozat standard jelöléssel, pl. R U R' U R U2 R'"
            >
              <input
                id="learn-notation"
                type="text"
                value={notation}
                onChange={(e) => setNotation(e.target.value)}
                maxLength={MAX_NOTATION_LENGTH}
                placeholder="R U R' U R U2 R'"
                className={`${controlClass} tabular font-mono`}
              />
            </Field>
          )}

          <Field label="Leírás" htmlFor="learn-description" optional>
            <textarea
              id="learn-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={MAX_DESCRIPTION_LENGTH}
              rows={4}
              placeholder="Mikor használd, mire figyelj…"
              className={controlClass}
            />
          </Field>

          {state.kind === "algorithm" && (
            <Field
              label="Videó URL"
              htmlFor="learn-video"
              optional
              hint="YouTube-linket beágyazva jelenítünk meg. Üresen hagyva nincs videó."
            >
              <input
                id="learn-video"
                type="url"
                inputMode="url"
                value={videoUrl}
                onChange={(e) => setVideoUrl(e.target.value)}
                maxLength={MAX_VIDEO_URL_LENGTH}
                placeholder="https://www.youtube.com/watch?v=…"
                className={controlClass}
              />
            </Field>
          )}

          <Field
            label="Sorrend"
            htmlFor="learn-order"
            hint="A megjelenítés sorrendje növekvő szerint; azonos érték esetén név szerint."
          >
            <input
              id="learn-order"
              type="number"
              step={1}
              value={order}
              onChange={(e) => setOrder(e.target.value)}
              className={`${controlClass} tabular font-mono`}
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
              {submitting ? "Mentés…" : isEdit ? "Mentés" : "Létrehozás"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
