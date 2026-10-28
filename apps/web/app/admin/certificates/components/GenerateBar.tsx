"use client";

import { useState } from "react";
import { FaChevronLeft, FaChevronRight, FaFileCsv, FaFilePdf, FaSpinner } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import Modal from "@/app/components/ui/Modal";
import type { CertificateCompetition, CertificateMode, Recipient } from "../types";
import { downloadCsv, recipientsCsv } from "../lib/csv";
import { downloadBlob } from "../lib/download";
import { certificateContexts, printOrder } from "../lib/pdf/content";
import type { CertificateTemplate } from "../lib/template/model";
import type { CertificateContext } from "../lib/template/variables";
import TemplatePreview from "./TemplatePreview";

const FILE_NAMES: Record<CertificateMode, string> = {
  podium: "dobogosok",
  competitors: "versenyzok",
  registrants: "regisztraltak",
};

type Job = "all" | "sample";

type Outcome = {
  pages: number;
  overflowing: string[];
  unsupported: string[];
  qrFailed: string[];
  missingImages: number;
};

function contextName(context: CertificateContext | null): string {
  if (!context) return "—";
  const { values } = context;
  const name = values["competitor.name"] ?? "";
  return values["result.event"] ? `${name} · ${values["result.event"]} ${values["result.place"] ?? ""}`.trim() : name;
}

export default function GenerateBar({
  template,
  competition,
  mode,
  selected,
  total,
}: {
  template: CertificateTemplate;
  competition: CertificateCompetition;
  mode: CertificateMode;
  selected: Recipient[];
  total: number;
}) {
  const [running, setRunning] = useState<Job | null>(null);
  const [outcome, setOutcome] = useState<Outcome | null>(null);
  const [failed, setFailed] = useState(false);
  const [previewing, setPreviewing] = useState(false);
  const [previewIndex, setPreviewIndex] = useState(0);

  const contexts = certificateContexts(competition, mode, selected);
  const shownIndex = Math.min(previewIndex, Math.max(contexts.length - 1, 0));
  const previewContext = contexts[shownIndex] ?? null;
  const step = (delta: number) => setPreviewIndex((shownIndex + delta + contexts.length) % contexts.length);
  const baseName = `${competition.id}-${FILE_NAMES[mode]}`;

  const generate = async (job: Job) => {
    const recipients = job === "sample" ? printOrder(mode, selected).slice(0, 1) : selected;
    setRunning(job);
    setFailed(false);
    try {
      const { generateCertificates } = await import("../lib/pdf/generate");
      const result = await generateCertificates(template.document, competition, mode, recipients);
      downloadBlob(
        job === "sample" ? `oklevel-minta-${baseName}.pdf` : `oklevelek-${baseName}.pdf`,
        new Blob([result.bytes.slice()], { type: "application/pdf" }),
      );
      setOutcome({
        pages: recipients.length,
        overflowing: result.overflowing,
        unsupported: result.unsupported,
        qrFailed: result.qrFailed,
        missingImages: result.missingImages,
      });
    } catch (error) {
      console.error("Az oklevél-PDF nem készült el:", error);
      setFailed(true);
    } finally {
      setRunning(null);
    }
  };

  const busy = running !== null;
  const disabled = busy || selected.length === 0;

  return (
    <div className="sticky bottom-4 z-10 flex flex-col gap-3 rounded-card border border-line bg-surface p-4 shadow-panel">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-sm font-semibold text-foreground">
            <span className="tabular font-mono">{selected.length}</span> oklevél kijelölve
            <span className="font-normal text-muted-foreground">
              {" "}
              / <span className="tabular font-mono">{total}</span>
            </span>
          </p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            {mode === "podium"
              ? "Sorrend: versenyszámonként 3. → 2. → 1. hely, a díjátadás sorrendjében."
              : "Sorrend: névsor."}
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <button
            type="button"
            onClick={() => setPreviewing(true)}
            className="group flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
          >
            <span className="overflow-hidden rounded-sm border border-line transition-colors group-hover:border-line-strong">
              <TemplatePreview
                document={template.document}
                context={previewContext}
                label={`A(z) „${template.name}” sablon előnézete`}
                className="h-12"
              />
            </span>
            Előnézet
          </button>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="ghost"
              size="sm"
              disabled={busy || selected.length === 0}
              onClick={() => downloadCsv(`oklevel-${baseName}.csv`, recipientsCsv(competition, mode, selected))}
              icon={<FaFileCsv className="h-3.5 w-3.5" />}
            >
              CSV
            </Button>
            <Button
              variant="secondary"
              size="sm"
              disabled={disabled}
              onClick={() => generate("sample")}
              icon={running === "sample" ? <FaSpinner className="h-3 w-3 animate-spin" /> : <FaFilePdf className="h-3.5 w-3.5" />}
            >
              Minta
            </Button>
            <Button
              variant="primary"
              size="sm"
              disabled={disabled}
              onClick={() => generate("all")}
              icon={running === "all" ? <FaSpinner className="h-3 w-3 animate-spin" /> : <FaFilePdf className="h-3.5 w-3.5" />}
            >
              {running === "all" ? "PDF készül…" : "PDF letöltése"}
            </Button>
          </div>
        </div>
      </div>

      {previewing && (
        <Modal title={template.name} size="lg" onClose={() => setPreviewing(false)}>
          <div className="flex flex-col gap-3">
            <TemplatePreview
              document={template.document}
              context={previewContext}
              label={`A(z) „${template.name}” sablon előnézete`}
              className="w-full border border-line"
            />
            <div className="flex items-center justify-between gap-3">
              <p className="min-w-0 truncate text-sm text-foreground">
                {contexts.length === 0 ? (
                  <span className="text-muted-foreground">Nincs kijelölt címzett, az adatmezők nevei látszanak.</span>
                ) : (
                  contextName(previewContext)
                )}
              </p>
              {contexts.length > 1 && (
                <div className="flex shrink-0 items-center gap-1 rounded-control border border-line px-1">
                  <button type="button" aria-label="Előző címzett" onClick={() => step(-1)} className="flex h-7 w-6 items-center justify-center text-muted-foreground hover:text-foreground">
                    <FaChevronLeft className="h-3 w-3" />
                  </button>
                  <span className="tabular font-mono text-xs text-muted-foreground">
                    {shownIndex + 1}/{contexts.length}
                  </span>
                  <button type="button" aria-label="Következő címzett" onClick={() => step(1)} className="flex h-7 w-6 items-center justify-center text-muted-foreground hover:text-foreground">
                    <FaChevronRight className="h-3 w-3" />
                  </button>
                </div>
              )}
            </div>
          </div>
        </Modal>
      )}

      {failed && (
        <p className="text-sm text-brand-red">
          A PDF nem készült el. Próbáld újra; ha ismét hibát kapsz, a böngésző konzolja megmutatja az okát.
        </p>
      )}
      {!failed && outcome && (
        <div className="flex flex-col gap-1 border-t border-line pt-3 text-sm">
          <p className="text-muted-foreground">
            Letöltve: <span className="tabular font-mono">{outcome.pages}</span> oldal.
          </p>
          {outcome.overflowing.length > 0 && (
            <p className="text-brand-red">
              Ezeknél valamelyik szöveg a legkisebb betűmérettel sem fér a helyére, nézd át őket:{" "}
              {outcome.overflowing.join(", ")}
            </p>
          )}
          {outcome.unsupported.length > 0 && (
            <p className="text-brand-red">
              Ezekben olyan karakter van, ami a betűkészletből hiányzik: {outcome.unsupported.join(", ")}
            </p>
          )}
          {outcome.qrFailed.length > 0 && (
            <p className="text-brand-red">
              Ezeknél a QR-kód nem készült el vagy túl kicsi a biztos beolvasáshoz: {outcome.qrFailed.join(", ")}
            </p>
          )}
          {outcome.missingImages > 0 && (
            <p className="text-brand-red">
              {outcome.missingImages} kép nem tölthető be, ezek kimaradtak a PDF-ből.
            </p>
          )}
        </div>
      )}
    </div>
  );
}
