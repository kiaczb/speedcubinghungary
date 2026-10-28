"use client";

import { useEffect, useState } from "react";
import { FaExclamationTriangle, FaSpinner } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import StateMessage from "@/app/components/ui/StateMessage";
import { ApiError, clientJson } from "@/app/lib/api";
import type { CertificateTemplate } from "../../../lib/template/model";
import TemplateEditor from "./TemplateEditor";

export default function TemplateLoader({ id }: { id: string }) {
  const [template, setTemplate] = useState<CertificateTemplate | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    clientJson<CertificateTemplate>(`/certificates/templates/${encodeURIComponent(id)}`)
      .then((loaded) => {
        if (!cancelled) setTemplate(loaded);
      })
      .catch((err) => {
        if (cancelled) return;
        console.error("A sablon nem tölthető be:", err);
        setError(
          err instanceof ApiError && (err.status === 404 || err.status === 400)
            ? "Ilyen sablon nincs (talán törölték)."
            : "A sablon most nem tölthető be. Próbáld újra kicsit később.",
        );
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  if (template) return <TemplateEditor key={template.id} template={template} />;

  return (
    <div className="px-5 py-16">
      {error ? (
        <StateMessage icon={<FaExclamationTriangle className="h-5 w-5" />} title="Nem sikerült" description={error}>
          <Button href="/admin/certificates" variant="secondary" size="sm">
            Vissza a sablonokhoz
          </Button>
        </StateMessage>
      ) : (
        <StateMessage icon={<FaSpinner className="h-5 w-5 animate-spin" />} title="Sablon betöltése…" description="Egy pillanat." />
      )}
    </div>
  );
}
