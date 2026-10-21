"use client";
import { useState } from "react";
import Button from "./Button";
import Modal from "./Modal";

type Props = {
  title: string;
  description: React.ReactNode;
  onConfirm: () => Promise<void>;
  onClose: () => void;
};

export default function ConfirmModal({
  title,
  description,
  onConfirm,
  onClose,
}: Props) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

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

  return (
    <Modal
      title={title}
      onClose={onClose}
      size="sm"
      role="alertdialog"
      closeDisabled={submitting}
      hideHeader
    >
      <div className="px-6 py-5">
        <h2 className="text-lg font-bold text-foreground">
          {title}
        </h2>
        <div className="mt-2 text-sm leading-relaxed text-muted-foreground">
          {description}
        </div>

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
            Mégse
          </Button>
          <Button
            type="button"
            variant="danger"
            onClick={handleConfirm}
            disabled={submitting}
          >
            {submitting ? "Törlés…" : "Törlés"}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
