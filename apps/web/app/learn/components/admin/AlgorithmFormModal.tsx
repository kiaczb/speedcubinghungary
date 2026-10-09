"use client";
import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";
import Modal from "@/app/components/ui/Modal";
import { clientFetch } from "@/app/lib/api";
import type { LearnAlgorithm } from "../../types";

type Props =
  | { mode: "create"; nodeId: string; nodeTitle: string; onClose: () => void }
  | {
      mode: "edit";
      algorithm: LearnAlgorithm;
      nodeTitle: string;
      onClose: () => void;
    };

export default function AlgorithmFormModal(props: Props) {
  const router = useRouter();
  const editing = props.mode === "edit" ? props.algorithm : null;

  const [notation, setNotation] = useState(editing?.notation ?? "");
  const [name, setName] = useState(editing?.name ?? "");
  const [notes, setNotes] = useState(editing?.notes ?? "");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const notationRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!notation.trim()) {
      setError("Add meg az algoritmus lépéseit.");
      return;
    }

    const body = { notation: notation.trim(), name: name.trim(), notes: notes.trim() };

    setSubmitting(true);
    try {
      if (props.mode === "create") {
        await clientFetch(`/learn/nodes/${props.nodeId}/algorithms`, {
          method: "POST",
          json: body,
        });
      } else {
        await clientFetch(`/learn/algorithms/${props.algorithm.id}`, {
          method: "PUT",
          json: body,
        });
      }
      router.refresh();
      props.onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={editing ? "Algoritmus szerkesztése" : `Új algoritmus — ${props.nodeTitle}`}
      onClose={props.onClose}
      size="lg"
      closeDisabled={submitting}
      initialFocusRef={notationRef}
    >
      <form onSubmit={handleSubmit} className="space-y-5 px-6 py-5">
        <Field
          label="Algoritmus"
          htmlFor="learn-algorithm-notation"
          hint="WCA-jelöléssel, szóközzel elválasztott lépésekkel, pl. R U R' U'. A sortöréseket és a dupla szóközöket egyetlen szóközzé alakítjuk."
        >
          <textarea
            id="learn-algorithm-notation"
            ref={notationRef}
            value={notation}
            onChange={(e) => setNotation(e.target.value)}
            rows={2}
            maxLength={500}
            spellCheck={false}
            autoCapitalize="off"
            autoCorrect="off"
            className={`${controlClass} font-mono`}
          />
        </Field>

        <Field
          label="Megnevezés"
          htmlFor="learn-algorithm-name"
          optional
          hint="Pl. „Fő”, „Balkezes”, „OH-barát”."
        >
          <input
            id="learn-algorithm-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={100}
            className={controlClass}
          />
        </Field>

        <Field
          label="Megjegyzés"
          htmlFor="learn-algorithm-notes"
          optional
          hint="Pl. ujjrend, felismerés, mikor érdemes ezt választani."
        >
          <textarea
            id="learn-algorithm-notes"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
            maxLength={2000}
            className={controlClass}
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
            onClick={props.onClose}
            disabled={submitting}
          >
            Mégse
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Mentés…" : editing ? "Mentés" : "Hozzáadás"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
