"use client";
import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { useBodyScrollLock } from "@/app/lib/useBodyScrollLock";
import Button from "./Button";

type Props = {
  title: string;
  description: React.ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => Promise<void>;
  onClose: () => void;
};


export default function ConfirmModal({
  title,
  description,
  confirmLabel = "Törlés",
  cancelLabel = "Mégse",
  onConfirm,
  onClose,
}: Props) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !submitting) onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onClose, submitting]);

  useBodyScrollLock(true);

  const handleConfirm = async () => {
    setError(null);
    setSubmitting(true);
    try {
      await onConfirm();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
      setSubmitting(false);
    }
  };

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm dark:bg-black/70"
      // A hátterre kattintás CSAK bezár (onClose), sose törli — a törlés
      // kizárólag a lenti "Törlés" gomb handleConfirm-jén keresztül fut.
      onClick={() => !submitting && onClose()}
      role="presentation"
    >
      <div
        className="w-full max-w-sm rounded-panel border border-line bg-surface shadow-panel"
        onClick={(e) => e.stopPropagation()}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-modal-title"
      >
        <div className="px-6 py-5">
          <h2 id="confirm-modal-title" className="text-lg font-bold text-foreground">
            {title}
          </h2>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
            {description}
          </p>

          {error && (
            <p
              role="alert"
              className="mt-4 rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong"
            >
              {error}
            </p>
          )}

          <div className="mt-5 flex justify-end gap-3">
            <Button
              type="button"
              variant="secondary"
              onClick={onClose}
              disabled={submitting}
            >
              {cancelLabel}
            </Button>
            <Button
              type="button"
              variant="danger"
              onClick={handleConfirm}
              disabled={submitting}
            >
              {submitting ? "Törlés…" : confirmLabel}
            </Button>
          </div>
        </div>
      </div>
    </div>,
    document.body,
  );
}
