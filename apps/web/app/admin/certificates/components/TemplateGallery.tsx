"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { FaExclamationTriangle, FaPlus, FaRegClone, FaRegTrashAlt, FaSpinner } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import ConfirmModal from "@/app/components/ui/ConfirmModal";
import StateMessage from "@/app/components/ui/StateMessage";
import { clientFetch, clientJson } from "@/app/lib/api";
import { formatDayMonthTime } from "@/app/lib/date";
import type { CertificateTemplate } from "../lib/template/model";
import { KIND_LABELS, pageLabel } from "../lib/template/labels";
import { sampleSet } from "../lib/template/samples";
import CreateTemplateModal from "./CreateTemplateModal";
import TemplatePreview from "./TemplatePreview";

const iconButtonClass =
  "flex h-8 w-8 items-center justify-center rounded-control text-muted-foreground transition-colors hover:bg-surface-muted hover:text-foreground disabled:opacity-40";

export function generateHref(templateId: string): string {
  return `/admin/certificates/generate?${new URLSearchParams({ template: templateId })}`;
}

function TemplateCard({
  template,
  copying,
  onCopy,
  onDelete,
}: {
  template: CertificateTemplate;
  copying: boolean;
  onCopy: () => void;
  onDelete: () => void;
}) {
  const { page } = template.document;
  const href = generateHref(template.id);
  return (
    <div className="flex flex-col overflow-hidden rounded-card border border-line bg-surface shadow-card transition-[box-shadow,border-color] duration-200 ease-smooth hover:border-line-strong hover:shadow-card-hover">
      <Link href={href} tabIndex={-1} aria-hidden className="flex h-48 items-center justify-center border-b border-line bg-surface-muted p-5">
        <TemplatePreview
          document={template.document}
          context={sampleSet(template.kind).contexts[0] ?? null}
          label={`A(z) „${template.name}” sablon előnézete`}
          className="h-full max-w-full shadow-card"
        />
      </Link>
      <div className="flex flex-1 flex-col gap-4 p-4">
        <div className="min-w-0">
          <Link href={href} className="block truncate font-semibold text-foreground hover:underline">
            {template.name}
          </Link>
          <p className="mt-1 text-sm text-muted-foreground">
            {KIND_LABELS[template.kind].short} · {pageLabel(page.width, page.height)}
          </p>
          <p className="tabular mt-0.5 font-mono text-xs text-subtle-foreground">
            {formatDayMonthTime(template.updatedAt)}
            {template.updatedByName && <span className="font-sans"> · {template.updatedByName}</span>}
          </p>
        </div>
        <div className="mt-auto flex items-center gap-2">
          <Button size="sm" href={href}>
            Használat
          </Button>
          <Button size="sm" variant="secondary" href={`/admin/certificates/templates/${template.id}`}>
            Szerkesztés
          </Button>
          <div className="ml-auto flex items-center">
            <button type="button" title="Másolat" aria-label={`„${template.name}” másolása`} disabled={copying} onClick={onCopy} className={iconButtonClass}>
              {copying ? <FaSpinner className="h-3.5 w-3.5 animate-spin" /> : <FaRegClone className="h-3.5 w-3.5" />}
            </button>
            <button type="button" title="Törlés" aria-label={`„${template.name}” törlése`} onClick={onDelete} className={iconButtonClass}>
              <FaRegTrashAlt className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function TemplateGallery() {
  const [templates, setTemplates] = useState<CertificateTemplate[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [reloads, setReloads] = useState(0);
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<CertificateTemplate | null>(null);
  const [copying, setCopying] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    clientJson<CertificateTemplate[]>("/certificates/templates")
      .then((data) => {
        if (cancelled) return;
        setTemplates(data);
        setFailed(false);
      })
      .catch((error) => {
        console.error("A sablonlista nem tölthető le:", error);
        if (!cancelled) setFailed(true);
      });
    return () => {
      cancelled = true;
    };
  }, [reloads]);

  const copy = async (template: CertificateTemplate) => {
    setCopying(template.id);
    try {
      await clientJson("/certificates/templates", {
        method: "POST",
        json: {
          name: `${template.name} (másolat)`.slice(0, 120),
          description: template.description,
          kind: template.kind,
          document: template.document,
        },
      });
      setReloads((n) => n + 1);
    } catch (error) {
      console.error("A sablon nem másolható:", error);
    } finally {
      setCopying(null);
    }
  };

  if (failed && !templates) {
    return (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="A sablonok nem tölthetők be"
        description="Próbáld újra kicsit később."
      />
    );
  }
  if (!templates) {
    return (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Sablonok betöltése…"
        description="Egy pillanat."
      />
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {templates.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Még nincs sablon. Kell egy dobogós a dobogósokhoz, és egy részvételi a versenyzőkhöz és a regisztráltakhoz.
        </p>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {templates.map((template) => (
          <TemplateCard
            key={template.id}
            template={template}
            copying={copying === template.id}
            onCopy={() => copy(template)}
            onDelete={() => setDeleting(template)}
          />
        ))}
        <button
          type="button"
          onClick={() => setCreating(true)}
          className="flex min-h-72 flex-col items-center justify-center gap-2 rounded-card border border-dashed border-line-strong text-sm font-medium text-muted-foreground transition-colors hover:border-foreground hover:bg-surface hover:text-foreground"
        >
          <FaPlus className="h-4 w-4" />
          Új sablon
        </button>
      </div>

      {creating && <CreateTemplateModal onClose={() => setCreating(false)} />}
      {deleting && (
        <ConfirmModal
          title="Sablon törlése"
          description={`Biztosan törlöd a(z) „${deleting.name}” sablont? Ez nem vonható vissza.`}
          onConfirm={async () => {
            await clientFetch(`/certificates/templates/${deleting.id}`, { method: "DELETE" });
            setReloads((n) => n + 1);
          }}
          onClose={() => setDeleting(null)}
        />
      )}
    </div>
  );
}
