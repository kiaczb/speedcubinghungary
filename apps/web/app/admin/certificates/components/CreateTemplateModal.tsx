"use client";

import { useState } from "react";
import type { FormEvent } from "react";
import { useRouter } from "next/navigation";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";
import Modal from "@/app/components/ui/Modal";
import { ApiError, clientJson } from "@/app/lib/api";
import { blankDocument } from "../lib/template/model";
import type { CertificateTemplate, TemplateKind } from "../lib/template/model";
import { KIND_LABELS } from "../lib/template/labels";

export default function CreateTemplateModal({ onClose }: { onClose: () => void }) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [kind, setKind] = useState<TemplateKind>("placement");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const created = await clientJson<CertificateTemplate>("/certificates/templates", {
        method: "POST",
        json: {
          name,
          description: "",
          kind,
          document: blankDocument(),
        },
      });
      router.push(`/admin/certificates/templates/${created.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "A sablon nem jött létre. Próbáld újra.");
      setSubmitting(false);
    }
  };

  return (
    <Modal title="Új sablon" onClose={onClose} size="md" closeDisabled={submitting}>
      <form onSubmit={submit} className="flex flex-col gap-5 px-6 py-5">
        <Field label="Név" htmlFor="template-name">
          <input
            id="template-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={120}
            required
            autoFocus
            placeholder="pl. Hungarian Open dobogós"
            className={controlClass}
          />
        </Field>

        <fieldset className="flex flex-col gap-2">
          <legend className="mb-1.5 text-sm font-medium text-foreground">Mire szól?</legend>
          {(["placement", "participation"] as TemplateKind[]).map((option) => (
            <label key={option} className="flex cursor-pointer items-start gap-3 rounded-control border border-line px-3 py-2.5 text-sm has-[:checked]:border-foreground has-[:checked]:bg-surface-muted">
              <input
                type="radio"
                name="kind"
                checked={kind === option}
                onChange={() => setKind(option)}
                className="mt-0.5 accent-foreground"
              />
              <span>
                <span className="block font-medium text-foreground">{KIND_LABELS[option].title}</span>
                <span className="block text-muted-foreground">{KIND_LABELS[option].description}</span>
              </span>
            </label>
          ))}
          <p className="text-xs text-muted-foreground">A típus később nem módosítható, de a sablon lemásolható.</p>
        </fieldset>

        <p className="text-sm text-muted-foreground">Üres A4 fekvő lappal indul; a méret a szerkesztőben átállítható.</p>

        {error && <p className="text-sm text-brand-red">{error}</p>}

        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose} disabled={submitting}>
            Mégse
          </Button>
          <Button type="submit" disabled={submitting || !name.trim()}>
            {submitting ? "Létrehozás…" : "Létrehozás és szerkesztés"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
