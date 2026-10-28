"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { FaArrowLeft, FaExclamationTriangle, FaSpinner } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import StateMessage from "@/app/components/ui/StateMessage";
import { ApiError, clientJson } from "@/app/lib/api";
import CompetitionPicker from "../../stats/components/CompetitionPicker";
import type { CertificateMode } from "../types";
import { parseMode } from "../lib/recipients";
import type { CertificateTemplate } from "../lib/template/model";
import { KIND_LABELS, pageLabel } from "../lib/template/labels";
import { sampleSet } from "../lib/template/samples";
import CompetitionCertificates from "./CompetitionCertificates";
import TemplatePreview from "./TemplatePreview";

const BASE_PATH = "/admin/certificates/generate";

type Loaded = { id: string; template: CertificateTemplate | null; error: string | null };

export default function GenerateView() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const templateId = searchParams.get("template") ?? "";
  const competitionId = searchParams.get("competition");
  const [loaded, setLoaded] = useState<Loaded | null>(null);

  useEffect(() => {
    if (!templateId) return;
    let cancelled = false;
    clientJson<CertificateTemplate>(`/certificates/templates/${encodeURIComponent(templateId)}`)
      .then((template) => {
        if (!cancelled) setLoaded({ id: templateId, template, error: null });
      })
      .catch((error) => {
        if (cancelled) return;
        console.error("A sablon nem tölthető be:", error);
        setLoaded({
          id: templateId,
          template: null,
          error:
            error instanceof ApiError && (error.status === 404 || error.status === 400)
              ? "Ilyen sablon nincs (talán törölték)."
              : "A sablon most nem tölthető be. Próbáld újra kicsit később.",
        });
      });
    return () => {
      cancelled = true;
    };
  }, [templateId]);

  const current = loaded?.id === templateId ? loaded : null;
  const template = current?.template ?? null;

  if (!templateId || current?.error) {
    return (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title={templateId ? "Nem sikerült" : "Nincs kiválasztott sablon"}
        description={current?.error ?? "Az oklevelekhez előbb válassz egy sablont."}
      >
        <Button href="/admin/certificates" variant="secondary" size="sm">
          Sablonok
        </Button>
      </StateMessage>
    );
  }
  if (!template) {
    return (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Sablon betöltése…"
        description="Egy pillanat."
      />
    );
  }

  const mode = parseMode(searchParams.get("mode"), template.kind);
  const open = (params: Record<string, string>) =>
    router.replace(`${BASE_PATH}?${new URLSearchParams({ template: template.id, ...params })}`);
  const { page } = template.document;

  return (
    <div className="flex flex-col gap-10">
      <div className="flex flex-col gap-6 border-b border-line pb-8">
        <div>
          <Button variant="ghost" size="sm" href="/admin/certificates" icon={<FaArrowLeft className="h-3 w-3" />}>
            Sablonok
          </Button>
        </div>
        <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
          <div className="flex h-28 shrink-0 items-center justify-start">
            <TemplatePreview
              document={template.document}
              context={sampleSet(template.kind).contexts[0] ?? null}
              label={`A(z) „${template.name}” sablon előnézete`}
              className="h-full border border-line shadow-card"
            />
          </div>
          <div className="min-w-0 flex-1">
            <h1 className="text-3xl font-bold text-foreground sm:text-4xl">{template.name}</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {KIND_LABELS[template.kind].short} · {pageLabel(page.width, page.height)}
            </p>
          </div>
          <Button variant="secondary" size="sm" href={`/admin/certificates/templates/${template.id}`}>
            Sablon szerkesztése
          </Button>
        </div>
      </div>

      {competitionId ? (
        <CompetitionCertificates
          key={competitionId}
          template={template}
          competitionId={competitionId}
          mode={mode}
          onModeChange={(nextMode: CertificateMode) => open({ competition: competitionId, mode: nextMode })}
          onBack={() => open({})}
        />
      ) : (
        <div className="flex flex-col gap-4">
          <h2 className="text-xl font-semibold text-foreground">Melyik versenyhez?</h2>
          <CompetitionPicker onSelect={(id) => open({ competition: id, mode })} />
        </div>
      )}
    </div>
  );
}
