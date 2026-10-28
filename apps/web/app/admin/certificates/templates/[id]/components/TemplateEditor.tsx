"use client";

import { useEffect, useEffectEvent, useReducer, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import {
  FaArrowLeft,
  FaChevronLeft,
  FaChevronRight,
  FaDatabase,
  FaExclamationTriangle,
  FaFilePdf,
  FaRedo,
  FaSave,
  FaSearchMinus,
  FaSearchPlus,
  FaSpinner,
  FaUndo,
} from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import ConfirmModal from "@/app/components/ui/ConfirmModal";
import Modal from "@/app/components/ui/Modal";
import { ApiError, clientFetch, clientJson } from "@/app/lib/api";
import { downloadBlob } from "../../../lib/download";
import type {
  CertificateAsset,
  CertificateDocument,
  CertificateElement,
  CertificateTemplate,
  PageSettings,
} from "../../../lib/template/model";
import { documentFontFiles } from "../../../lib/template/fonts";
import type { Measurer } from "../../../lib/template/textLayout";
import { elementSegments } from "../../../lib/template/model";
import { qrText } from "../../../lib/template/qr";
import { isKnownVariable, variableLabel } from "../../../lib/template/variables";
import { KIND_LABELS } from "../../../lib/template/labels";
import {
  editorReducer,
  initialState,
  insertElement,
  moveLayer,
  removeElement,
  updateElement,
} from "../lib/editorState";
import type { LayerMove } from "../lib/editorState";
import { backgroundImage, createElement, duplicate } from "../lib/elements";
import type { NewElement } from "../lib/elements";
import { assetObjectUrl, editorMeasurer } from "../../../lib/template/resources";
import { contextName, sampleSet } from "../../../lib/template/samples";
import type { PreviewSet } from "../../../lib/template/samples";
import CertificateCanvas from "./CertificateCanvas";
import { qrProblem, textLayouts } from "../../../components/CertificateShapes";
import EditorSidebar from "./EditorSidebar";
import PropertiesPanel from "./PropertiesPanel";
import RealDataModal from "./RealDataModal";
import type { TextContentEditorHandle } from "./TextContentEditor";

const PX_PER_MM = 96 / 25.4;
const ZOOM_STEPS = [25, 33, 50, 67, 75, 100, 125, 150, 200, 300, 400];

type Zoom = "fit" | number;

function isEditable(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLElement &&
    (target.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName))
  );
}

function scalePage(document: CertificateDocument, page: PageSettings, scaleElements: boolean): CertificateDocument {
  const sx = page.width / document.page.width;
  const sy = page.height / document.page.height;
  if (!scaleElements || (sx === 1 && sy === 1)) return { ...document, page };
  const r = (v: number) => Math.round(v * 10) / 10;
  const font = Math.min(sx, sy);
  return {
    ...document,
    page,
    elements: document.elements.map((element) => {
      const box = { x: r(element.x * sx), y: r(element.y * sy), width: r(element.width * sx), height: r(element.height * sy) };
      if (element.type !== "text") return { ...element, ...box };
      return {
        ...element,
        ...box,
        style: { ...element.style, size: r(element.style.size * font), minSize: r(element.style.minSize * font) },
      };
    }),
  };
}

export default function TemplateEditor({ template }: { template: CertificateTemplate }) {
  const router = useRouter();
  const [state, dispatch] = useReducer(editorReducer, template, initialState);
  const [measurer, setMeasurer] = useState<Measurer | null>(null);
  const [assets, setAssets] = useState<CertificateAsset[]>([]);
  const [assetUrls, setAssetUrls] = useState<Record<string, string>>({});
  const [preview, setPreview] = useState<PreviewSet>(() => sampleSet(template.kind));
  const [previewIndex, setPreviewIndex] = useState(0);
  const [showLabels, setShowLabels] = useState(false);
  const [zoom, setZoom] = useState<Zoom>("fit");
  const [viewport, setViewport] = useState({ width: 0, height: 0 });
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [conflict, setConflict] = useState(false);
  const [leaving, setLeaving] = useState(false);
  const [realData, setRealData] = useState(false);
  const [deletingAsset, setDeletingAsset] = useState<CertificateAsset | null>(null);
  const [testing, setTesting] = useState(false);
  const viewportRef = useRef<HTMLDivElement>(null);
  const contentEditorRef = useRef<TextContentEditorHandle>(null);
  const dragBefore = useRef<CertificateDocument | null>(null);
  const clipboard = useRef<CertificateElement | null>(null);

  const { document, selectedId } = state;
  const kind = state.template.kind;
  const selected = document.elements.find((e) => e.id === selectedId) ?? null;
  const context = showLabels ? null : (preview.contexts[previewIndex] ?? null);
  const place = context?.place ?? preview.contexts[previewIndex]?.place ?? 1;

  const fontKey = documentFontFiles(document).sort().join("|");
  useEffect(() => {
    let cancelled = false;
    editorMeasurer(fontKey ? fontKey.split("|") : [])
      .then((loaded) => {
        if (!cancelled) setMeasurer(loaded);
      })
      .catch((error) => console.error("A betűkészlet nem tölthető be:", error));
    return () => {
      cancelled = true;
    };
  }, [fontKey]);

  useEffect(() => {
    clientJson<CertificateAsset[]>("/certificates/assets")
      .then(setAssets)
      .catch((error) => console.error("A képtár nem tölthető le:", error));
  }, []);

  const assetKey = [
    ...new Set([
      ...assets.map((a) => a.id),
      ...document.elements.flatMap((e) => (e.type === "image" && e.assetId ? [e.assetId] : [])),
    ]),
  ]
    .sort()
    .join("|");
  useEffect(() => {
    let cancelled = false;
    for (const id of assetKey ? assetKey.split("|") : []) {
      assetObjectUrl(id)
        .then((url) => {
          if (!cancelled) setAssetUrls((current) => (current[id] === url ? current : { ...current, [id]: url }));
        })
        .catch((error) => console.error("A kép nem tölthető be:", id, error));
    }
    return () => {
      cancelled = true;
    };
  }, [assetKey]);

  useEffect(() => {
    const element = viewportRef.current;
    if (!element) return;
    const observer = new ResizeObserver(([entry]) => {
      setViewport({ width: entry.contentRect.width, height: entry.contentRect.height });
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!state.dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [state.dirty]);

  const fitScale = Math.max(
    0.5,
    Math.min((viewport.width - 64) / document.page.width, (viewport.height - 64) / document.page.height),
  );
  const scale = zoom === "fit" ? fitScale : (zoom / 100) * PX_PER_MM;
  const zoomPercent = Math.round((scale / PX_PER_MM) * 100);

  const change = (next: CertificateDocument, key?: string) =>
    dispatch({ type: "change", document: next, key, at: Date.now() });
  const updateSelected = (update: (element: CertificateElement) => CertificateElement, key: string) => {
    if (selectedId) change(updateElement(document, selectedId, update), key);
  };
  const add = (spec: NewElement, at?: { x: number; y: number }) => {
    const element = createElement(spec, at ?? { x: document.page.width / 2, y: document.page.height / 2 }, document.page);
    change(insertElement(document, element));
    dispatch({ type: "select", id: element.id });
  };
  const removeSelected = () => {
    if (!selectedId) return;
    change(removeElement(document, selectedId));
    dispatch({ type: "select", id: null });
  };
  const duplicateSelected = () => {
    if (!selected) return;
    const copy = duplicate(selected);
    change(insertElement(document, copy, document.elements.indexOf(selected) + 1));
    dispatch({ type: "select", id: copy.id });
  };
  const layer = (move: LayerMove) => selectedId && change(moveLayer(document, selectedId, move));

  const save = async (overrideVersion?: number) => {
    if (saving) return;
    if (!state.name.trim()) {
      setSaveError("A sablonnak nevet kell adni (oldal-tulajdonságok, felül).");
      return;
    }
    setSaving(true);
    setSaveError(null);
    const snapshot = { document, name: state.name, description: state.description };
    try {
      const saved = await clientJson<CertificateTemplate>(`/certificates/templates/${state.template.id}`, {
        method: "PUT",
        json: { ...snapshot, version: overrideVersion ?? state.template.version },
      });
      dispatch({ type: "saved", template: saved, ...snapshot });
      setConflict(false);
    } catch (error) {
      if (error instanceof ApiError && error.status === 409) setConflict(true);
      else if (error instanceof ApiError && error.status === 400) setSaveError(`Nem menthető: ${error.message}`);
      else setSaveError("A mentés nem sikerült. A változtatásaid még itt vannak; próbáld újra.");
    } finally {
      setSaving(false);
    }
  };

  const loadServerVersion = async () => {
    const latest = await clientJson<CertificateTemplate>(`/certificates/templates/${state.template.id}`);
    return latest;
  };

  const upload = async (file: File) => {
    const body = new FormData();
    body.append("file", file);
    try {
      const asset = await clientJson<CertificateAsset>("/certificates/assets", { method: "POST", body });
      setAssets((current) => [asset, ...current]);
    } catch (error) {
      if (error instanceof ApiError && error.status === 400) {
        throw new Error("Csak PNG vagy JPEG kép tölthető fel, legfeljebb 10 MB.");
      }
      throw new Error("A feltöltés nem sikerült.");
    }
  };

  const testPdf = async () => {
    setTesting(true);
    try {
      const [{ renderCertificates }, loaders] = await Promise.all([
        import("../../../lib/pdf/render"),
        import("../../../lib/template/loaders"),
      ]);
      const [fonts, assetBytes] = await Promise.all([loaders.loadDocumentFonts(document), loaders.loadDocumentAssets(document)]);
      const contexts = [preview.contexts[previewIndex]].filter(Boolean);
      const result = await renderCertificates(document, contexts, fonts, assetBytes, {
        title: `Próba – ${state.name}`,
        author: "Speedcubing Hungary",
      });
      downloadBlob("oklevel-sablon-proba.pdf", new Blob([result.bytes.slice()], { type: "application/pdf" }));
    } catch (error) {
      console.error("A próba-PDF nem készült el:", error);
      setSaveError("A próba-PDF nem készült el.");
    } finally {
      setTesting(false);
    }
  };

  const onKey = useEffectEvent((event: KeyboardEvent) => {
    const mod = event.ctrlKey || event.metaKey;
    const key = event.key.toLowerCase();
    if (mod && key === "s") {
      event.preventDefault();
      save();
      return;
    }
    if (isEditable(event.target) || leaving || realData) return;
    if (mod && key === "z") {
      event.preventDefault();
      dispatch({ type: event.shiftKey ? "redo" : "undo" });
    } else if (mod && key === "y") {
      event.preventDefault();
      dispatch({ type: "redo" });
    } else if (mod && key === "d") {
      event.preventDefault();
      duplicateSelected();
    } else if (mod && key === "c" && selected) {
      clipboard.current = selected;
    } else if (mod && key === "v" && clipboard.current) {
      const copy = duplicate(clipboard.current, 8);
      change(insertElement(document, copy));
      dispatch({ type: "select", id: copy.id });
    } else if ((event.key === "Delete" || event.key === "Backspace") && selected && !selected.locked) {
      event.preventDefault();
      removeSelected();
    } else if (event.key === "Escape") {
      dispatch({ type: "select", id: null });
    } else if (event.key.startsWith("Arrow") && selected && !selected.locked) {
      event.preventDefault();
      const step = event.shiftKey ? 10 : 1;
      const dx = event.key === "ArrowLeft" ? -step : event.key === "ArrowRight" ? step : 0;
      const dy = event.key === "ArrowUp" ? -step : event.key === "ArrowDown" ? step : 0;
      updateSelected((e) => ({ ...e, x: Math.round((e.x + dx) * 10) / 10, y: Math.round((e.y + dy) * 10) / 10 }), `${selected.id}:nudge`);
    }
  });

  useEffect(() => {
    const listener = (event: KeyboardEvent) => onKey(event);
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, []);

  const layouts = showLabels ? new Map() : textLayouts(document.elements.filter((e) => !e.hidden), context, measurer);
  const issues: string[] = [];
  for (const element of document.elements) {
    for (const segment of elementSegments(element)) {
      if ("variable" in segment && !isKnownVariable(segment.variable, kind)) {
        issues.push(`„${element.name}”: a(z) „${variableLabel(segment.variable)}” adatmező ${KIND_LABELS[kind].short.toLowerCase()} oklevélen üres marad.`);
      }
    }
    const qrIssue = element.hidden ? null : qrProblem(element, context);
    if (qrIssue) issues.push(`„${element.name}”: ${qrIssue}.`);
    if (element.type === "qr" && !element.hidden && context && qrText(element, context) === null) {
      issues.push(`„${element.name}” ennél a címzettnél nem jelenik meg (hiányzó adat, nincs tartalék): ${contextName(context)}.`);
    }
    if (element.type === "image" && !element.assetId) issues.push(`„${element.name}”: nincs kép kiválasztva.`);
    const layout = layouts.get(element.id);
    if (layout?.overflow) issues.push(`„${element.name}” nem fér el ezzel az adattal: ${contextName(context ?? undefined)}.`);
    else if (layout && !layout.covered) issues.push(`„${element.name}”: a betűkészletből hiányzó karakter (${contextName(context ?? undefined)}).`);
  }
  const selectedLayout = selected ? layouts.get(selected.id) : undefined;

  const status = saving ? "Mentés…" : state.dirty ? "Nem mentett változások" : "Mentve";

  return (
    <div className="flex h-[calc(100dvh-4rem-1px)] min-h-[36rem] flex-col">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-2 border-b border-line bg-surface px-4 py-2">
        <Button
          variant="ghost"
          size="sm"
          icon={<FaArrowLeft className="h-3 w-3" />}
          onClick={() => (state.dirty ? setLeaving(true) : router.push("/admin/certificates"))}
        >
          Sablonok
        </Button>
        <div className="flex min-w-0 items-center gap-2">
          <span className="truncate font-semibold text-foreground">{state.name || "Névtelen sablon"}</span>
          <span className="text-sm text-muted-foreground">{KIND_LABELS[kind].short}</span>
          <span className={`text-xs ${state.dirty ? "text-brand-red" : "text-subtle-foreground"}`}>{status}</span>
        </div>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          <div className="flex items-center rounded-control border border-line">
            <button type="button" title="Visszavonás (Ctrl+Z)" aria-label="Visszavonás" disabled={state.past.length === 0} onClick={() => dispatch({ type: "undo" })} className="flex h-8 w-8 items-center justify-center text-muted-foreground hover:text-foreground disabled:opacity-30">
              <FaUndo className="h-3 w-3" />
            </button>
            <button type="button" title="Újra (Ctrl+Y)" aria-label="Újra" disabled={state.future.length === 0} onClick={() => dispatch({ type: "redo" })} className="flex h-8 w-8 items-center justify-center text-muted-foreground hover:text-foreground disabled:opacity-30">
              <FaRedo className="h-3 w-3" />
            </button>
          </div>

          <div className="flex items-center rounded-control border border-line">
            <button
              type="button"
              aria-label="Kicsinyítés"
              onClick={() => setZoom(ZOOM_STEPS.filter((z) => z < zoomPercent).at(-1) ?? ZOOM_STEPS[0])}
              className="flex h-8 w-8 items-center justify-center text-muted-foreground hover:text-foreground"
            >
              <FaSearchMinus className="h-3 w-3" />
            </button>
            <button type="button" onClick={() => setZoom("fit")} title="Illesztés az ablakhoz" className="tabular w-14 font-mono text-xs text-foreground">
              {zoomPercent}%
            </button>
            <button
              type="button"
              aria-label="Nagyítás"
              onClick={() => setZoom(ZOOM_STEPS.find((z) => z > zoomPercent) ?? ZOOM_STEPS.at(-1)!)}
              className="flex h-8 w-8 items-center justify-center text-muted-foreground hover:text-foreground"
            >
              <FaSearchPlus className="h-3 w-3" />
            </button>
          </div>

          <div className="flex items-center rounded-control border border-line p-0.5">
            {[
              { labels: true, text: "Mezőnevek" },
              { labels: false, text: "Adatokkal" },
            ].map((option) => (
              <button
                key={option.text}
                type="button"
                aria-pressed={showLabels === option.labels}
                onClick={() => setShowLabels(option.labels)}
                className={`rounded-md px-2.5 py-1 text-xs font-medium ${showLabels === option.labels ? "bg-foreground text-background" : "text-muted-foreground hover:text-foreground"}`}
              >
                {option.text}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-1 rounded-control border border-line px-1">
            <button type="button" aria-label="Előző minta" disabled={showLabels} onClick={() => setPreviewIndex((i) => (i - 1 + preview.contexts.length) % preview.contexts.length)} className="flex h-7 w-6 items-center justify-center text-muted-foreground hover:text-foreground disabled:opacity-30">
              <FaChevronLeft className="h-3 w-3" />
            </button>
            <span className="flex min-w-0 items-baseline gap-1 text-xs">
              <span className="max-w-48 truncate text-foreground" title={showLabels ? undefined : contextName(preview.contexts[previewIndex])}>
                {showLabels ? "—" : contextName(preview.contexts[previewIndex])}
              </span>
              <span className="tabular shrink-0 font-mono text-subtle-foreground">
                {previewIndex + 1}/{preview.contexts.length}
              </span>
            </span>
            <button type="button" aria-label="Következő minta" disabled={showLabels} onClick={() => setPreviewIndex((i) => (i + 1) % preview.contexts.length)} className="flex h-7 w-6 items-center justify-center text-muted-foreground hover:text-foreground disabled:opacity-30">
              <FaChevronRight className="h-3 w-3" />
            </button>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setRealData(true)}
            title={preview.label === "Mintaadat" ? undefined : `Valódi adat: ${preview.label}`}
            className="max-w-56 whitespace-nowrap"
            icon={<FaDatabase className="h-3 w-3 shrink-0" />}
          >
            <span className="truncate">{preview.label === "Mintaadat" ? "Valódi adat betöltése" : preview.label}</span>
          </Button>
          <Button variant="secondary" size="sm" disabled={testing} onClick={testPdf} icon={testing ? <FaSpinner className="h-3 w-3 animate-spin" /> : <FaFilePdf className="h-3 w-3" />}>
            PDF-próba
          </Button>
          <Button size="sm" disabled={saving || !state.dirty} onClick={() => save()} icon={<FaSave className="h-3 w-3" />}>
            Mentés
          </Button>
        </div>
      </div>

      {(saveError || conflict || issues.length > 0) && (
        <div className="flex flex-col gap-1 border-b border-line bg-surface-muted px-4 py-2 text-sm">
          {conflict && (
            <div className="flex flex-wrap items-center gap-2 text-brand-red">
              <FaExclamationTriangle className="h-3.5 w-3.5" />
              Közben valaki más elmentette ezt a sablont.
              <Button size="sm" variant="secondary" onClick={async () => dispatch({ type: "replace", template: await loadServerVersion() })}>
                Az övé betöltése (az enyém elvész)
              </Button>
              <Button size="sm" variant="danger" onClick={async () => save((await loadServerVersion()).version)}>
                Felülírom az enyémmel
              </Button>
            </div>
          )}
          {saveError && <p className="text-brand-red">{saveError}</p>}
          {issues.slice(0, 4).map((issue) => (
            <p key={issue} className="text-muted-foreground">
              <FaExclamationTriangle className="mr-1.5 inline h-3 w-3 text-brand-red" />
              {issue}
            </p>
          ))}
          {issues.length > 4 && <p className="text-xs text-subtle-foreground">…és még {issues.length - 4} figyelmeztetés.</p>}
        </div>
      )}

      <div className="hidden min-h-0 flex-1 lg:grid lg:grid-cols-[16rem_minmax(0,1fr)_21rem]">
        <aside className="min-h-0 overflow-y-auto border-r border-line bg-surface">
          <EditorSidebar
            kind={kind}
            elements={document.elements}
            selectedId={selectedId}
            assets={assets}
            assetUrls={assetUrls}
            onAdd={(spec) => add(spec)}
            onSelect={(id) => dispatch({ type: "select", id })}
            onToggle={(id, flag) => change(updateElement(document, id, (e) => ({ ...e, [flag]: !e[flag] })))}
            onUpload={upload}
            onDeleteAsset={setDeletingAsset}
          />
        </aside>

        <div ref={viewportRef} className="min-h-0 overflow-auto bg-canvas" onPointerDown={() => dispatch({ type: "select", id: null })}>
          <div className="flex min-h-full min-w-max items-center justify-center p-8">
            <div onPointerDown={(event) => event.stopPropagation()}>
              <CertificateCanvas
                document={document}
                context={context}
                place={place}
                measurer={measurer}
                assetUrls={assetUrls}
                scale={scale}
                selectedId={selectedId}
                onSelect={(id) => dispatch({ type: "select", id })}
                onEditText={(id) => {
                  dispatch({ type: "select", id });
                  setTimeout(() => contentEditorRef.current?.focus(), 0);
                }}
                onTransformStart={() => {
                  dragBefore.current = document;
                }}
                onTransform={(id, box) =>
                  dispatch({ type: "preview", document: updateElement(document, id, (e) => ({ ...e, ...box })) })
                }
                onTransformEnd={() => {
                  if (dragBefore.current) dispatch({ type: "commit", before: dragBefore.current });
                  dragBefore.current = null;
                }}
                onDropElement={(data, at) => {
                  try {
                    add(JSON.parse(data) as NewElement, at);
                  } catch {}
                }}
              />
            </div>
          </div>
        </div>

        <aside className="min-h-0 overflow-y-auto border-l border-line bg-surface">
          <PropertiesPanel
            kind={kind}
            document={document}
            element={selected}
            overflow={Boolean(selectedLayout && (selectedLayout.overflow || !selectedLayout.covered))}
            qrPreview={selected?.type === "qr" ? qrText(selected, preview.contexts[previewIndex] ?? null) : null}
            assets={assets}
            assetUrls={assetUrls}
            name={state.name}
            description={state.description}
            contentEditorRef={contentEditorRef}
            onRename={(name) => dispatch({ type: "rename", name })}
            onDescribe={(description) => dispatch({ type: "describe", description })}
            onUpdate={updateSelected}
            onPage={(page, scaleElements) => change(scalePage(document, page, scaleElements), "page")}
            onLayer={layer}
            onDuplicate={duplicateSelected}
            onDelete={removeSelected}
            onBackground={(asset) => {
              const element = backgroundImage(asset.id, document.page);
              change(insertElement(document, element, 0));
              dispatch({ type: "select", id: element.id });
            }}
            onUpload={upload}
          />
        </aside>
      </div>

      <div className="flex flex-1 items-center justify-center p-8 text-center text-sm text-muted-foreground lg:hidden">
        A sablonszerkesztőhöz legalább 1024 képpont széles ablak kell. Nyisd meg számítógépen.
      </div>

      {realData && (
        <RealDataModal
          kind={kind}
          onClose={() => setRealData(false)}
          onLoaded={(set) => {
            setPreview(set);
            setPreviewIndex(0);
            setShowLabels(false);
          }}
        />
      )}

      {leaving && (
        <Modal title="Nem mentett változások" onClose={() => setLeaving(false)} size="sm">
          <div className="flex flex-col gap-4 px-6 py-5">
            <p className="text-sm text-muted-foreground">A legutóbbi mentés óta változtattál a sablonon. Mi legyen velük?</p>
            <div className="flex flex-wrap justify-end gap-2">
              <Button variant="ghost" onClick={() => setLeaving(false)}>
                Maradok
              </Button>
              <Button variant="danger" onClick={() => router.push("/admin/certificates")}>
                Elvetem
              </Button>
              <Button
                onClick={async () => {
                  await save();
                  setLeaving(false);
                }}
              >
                Mentés
              </Button>
            </div>
          </div>
        </Modal>
      )}

      {deletingAsset && (
        <ConfirmModal
          title="Kép törlése a képtárból"
          description={`Biztosan törlöd a(z) „${deletingAsset.name}” képet? Ha valamelyik sablon használja, nem törölhető.`}
          onConfirm={async () => {
            try {
              await clientFetch(`/certificates/assets/${deletingAsset.id}`, { method: "DELETE" });
            } catch (error) {
              if (error instanceof ApiError && error.status === 409) {
                throw new Error("Ezt a képet egy mentett sablon használja, ezért nem törölhető.");
              }
              throw error;
            }
            setAssets((current) => current.filter((a) => a.id !== deletingAsset.id));
          }}
          onClose={() => setDeletingAsset(null)}
        />
      )}
    </div>
  );
}
