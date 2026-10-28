"use client";

import { useState } from "react";
import type { Ref } from "react";
import {
  FaAlignCenter,
  FaAlignLeft,
  FaAlignRight,
  FaBold,
  FaItalic,
  FaLock,
  FaLockOpen,
  FaRegClone,
  FaRegEye,
  FaRegEyeSlash,
  FaRegTrashAlt,
  FaUnderline,
} from "react-icons/fa";
import {
  MdFlipToBack,
  MdFlipToFront,
  MdVerticalAlignBottom,
  MdVerticalAlignCenter,
  MdVerticalAlignTop,
} from "react-icons/md";
import { PAGE_PRESETS } from "../../../lib/template/model";
import type {
  CertificateAsset,
  CertificateDocument,
  CertificateElement,
  PageSettings,
  QrElement,
  TemplateKind,
  TextElement,
  TextSegment,
} from "../../../lib/template/model";
import { FONT_FAMILIES, supportsBold, supportsItalic } from "../../../lib/template/fonts";
import { KIND_LABELS } from "../../../lib/template/labels";
import type { LayerMove } from "../lib/editorState";
import { ELEMENT_LABELS, QR_PRESETS } from "../lib/elements";
import SegmentsField, { VariableSelect } from "./SegmentsField";
import AssetLibrary from "./AssetLibrary";
import TextContentEditor from "./TextContentEditor";
import type { TextContentEditorHandle } from "./TextContentEditor";
import { Checkbox, ColorInput, NumberInput, Row, Section, Select, TextInput, ToggleGroup } from "./controls";

type Update = (update: (element: CertificateElement) => CertificateElement, key: string) => void;

type Props = {
  kind: TemplateKind;
  document: CertificateDocument;
  element: CertificateElement | null;
  overflow: boolean;
  qrPreview: string | null;
  assets: CertificateAsset[];
  assetUrls: Record<string, string>;
  name: string;
  description: string;
  contentEditorRef: Ref<TextContentEditorHandle>;
  onRename: (name: string) => void;
  onDescribe: (description: string) => void;
  onUpdate: Update;
  onPage: (page: PageSettings, scaleElements: boolean) => void;
  onLayer: (move: LayerMove) => void;
  onDuplicate: () => void;
  onDelete: () => void;
  onBackground: (asset: CertificateAsset) => void;
  onUpload: (file: File) => Promise<void>;
};

function TextProperties({
  element,
  kind,
  overflow,
  onUpdate,
  contentEditorRef,
}: {
  element: TextElement;
  kind: TemplateKind;
  overflow: boolean;
  onUpdate: Update;
  contentEditorRef: Ref<TextContentEditorHandle>;
}) {
  const style = element.style;
  const set = (patch: Partial<TextElement["style"]>, key: string) =>
    onUpdate((e) => (e.type === "text" ? { ...e, style: { ...e.style, ...patch } } : e), `${element.id}:${key}`);
  const insertTarget = contentEditorRef && typeof contentEditorRef === "object" ? contentEditorRef : null;

  return (
    <>
      <Section title="Tartalom">
        <TextContentEditor
          key={element.id}
          id="text-content"
          ref={contentEditorRef}
          segments={element.content}
          onChange={(content) => onUpdate((e) => (e.type === "text" ? { ...e, content } : e), `${element.id}:content`)}
        />
        <VariableSelect kind={kind} onPick={(key) => insertTarget?.current?.insertVariable(key)} />
        {overflow && (
          <p className="text-xs text-brand-red">
            A mostani előnézeti adattal a szöveg nem fér el a dobozban (vagy hiányzó karakter van benne). Növeld a dobozt,
            vagy válassz „kicsinyítés” illeszkedést.
          </p>
        )}
      </Section>

      <Section title="Betű">
        <Select
          value={style.font}
          options={FONT_FAMILIES.map((f) => ({ value: f.id, label: f.label }))}
          onChange={(font) => set({ font, bold: style.bold && supportsBold(font), italic: style.italic && supportsItalic(font) }, "font")}
        />
        <div className="flex flex-wrap items-center gap-2">
          <div className="w-24">
            <NumberInput value={style.size} min={4} max={200} step={1} decimals={1} unit="pt" ariaLabel="Betűméret" onChange={(size) => set({ size }, "size")} />
          </div>
          <ToggleGroup
            label="Betűstílus"
            value={[style.bold ? "bold" : "", style.italic ? "italic" : "", style.underline ? "underline" : ""].filter(Boolean) as ("bold" | "italic" | "underline")[]}
            options={[
              { value: "bold", label: <FaBold className="h-3 w-3" />, title: "Félkövér", disabled: !supportsBold(style.font) },
              { value: "italic", label: <FaItalic className="h-3 w-3" />, title: "Dőlt", disabled: !supportsItalic(style.font) },
              { value: "underline", label: <FaUnderline className="h-3 w-3" />, title: "Aláhúzott" },
            ]}
            onChange={(flag) => set({ [flag]: !style[flag] }, flag)}
          />
          <ToggleGroup
            label="Nagybetűs"
            value={style.uppercase ? "upper" : ""}
            options={[{ value: "upper", label: "AA", title: "Csupa nagybetű" }]}
            onChange={() => set({ uppercase: !style.uppercase }, "uppercase")}
          />
        </div>
        <ColorInput value={style.color} allowMedal={kind === "placement"} onChange={(color) => set({ color: color ?? "#16181d" }, "color")} />
      </Section>

      <Section title="Elrendezés a dobozban">
        <div className="flex flex-wrap gap-2">
          <ToggleGroup
            label="Vízszintes igazítás"
            value={style.align}
            options={[
              { value: "left", label: <FaAlignLeft className="h-3 w-3" />, title: "Balra" },
              { value: "center", label: <FaAlignCenter className="h-3 w-3" />, title: "Középre" },
              { value: "right", label: <FaAlignRight className="h-3 w-3" />, title: "Jobbra" },
            ]}
            onChange={(align) => set({ align }, "align")}
          />
          <ToggleGroup
            label="Függőleges igazítás"
            value={style.verticalAlign}
            options={[
              { value: "top", label: <MdVerticalAlignTop className="h-4 w-4" />, title: "Fent" },
              { value: "middle", label: <MdVerticalAlignCenter className="h-4 w-4" />, title: "Középen" },
              { value: "bottom", label: <MdVerticalAlignBottom className="h-4 w-4" />, title: "Lent" },
            ]}
            onChange={(verticalAlign) => set({ verticalAlign }, "valign")}
          />
        </div>
        <Row label="Sorköz">
          <NumberInput value={style.lineHeight} min={0.8} max={3} step={0.05} decimals={2} unit="×" onChange={(lineHeight) => set({ lineHeight }, "lineHeight")} />
        </Row>
        <Row label="Betűköz">
          <NumberInput value={style.letterSpacing} min={-2} max={30} step={0.5} decimals={1} unit="pt" onChange={(letterSpacing) => set({ letterSpacing }, "letterSpacing")} />
        </Row>
        <Row label="Ha nem fér el">
          <Select
            value={style.fit}
            options={[
              { value: "wrap", label: "Sortörés" },
              { value: "shrink", label: "Kicsinyítés egy sorba" },
              { value: "wrapShrink", label: "Sortörés + kicsinyítés" },
            ]}
            onChange={(fit) => set({ fit }, "fit")}
          />
        </Row>
        {style.fit !== "wrap" && (
          <Row label="Legkisebb">
            <NumberInput value={style.minSize} min={4} max={style.size} step={1} unit="pt" onChange={(minSize) => set({ minSize }, "minSize")} />
          </Row>
        )}
      </Section>
    </>
  );
}

function QrProperties({
  element,
  kind,
  preview,
  set,
}: {
  element: QrElement;
  kind: TemplateKind;
  preview: string | null;
  set: (patch: Partial<QrElement>, key: string) => void;
}) {
  const same = (a: TextSegment[], b: TextSegment[]) => JSON.stringify(a) === JSON.stringify(b);
  const preset = QR_PRESETS.find((p) => same(p.content, element.content));
  const [custom, setCustom] = useState(!preset);
  const contentMode = custom || !preset ? "custom" : preset.id;
  const fallbackMode =
    element.fallback.length === 0
      ? "hide"
      : same(element.fallback, [{ variable: "competition.url" }])
        ? "competition"
        : "custom";

  return (
    <>
      <Section title="QR-kód tartalma">
        <Select
          value={contentMode}
          options={[...QR_PRESETS.map((p) => ({ value: p.id, label: p.label })), { value: "custom", label: "Egyéni cím vagy szöveg" }]}
          onChange={(id) => {
            const next = QR_PRESETS.find((p) => p.id === id);
            setCustom(!next);
            if (next) set({ content: next.content, fallback: next.fallback }, "qr-preset");
          }}
        />
        {contentMode === "custom" && (
          <SegmentsField id={`${element.id}-qr-content`} kind={kind} segments={element.content} onChange={(content) => set({ content }, "qr-content")} />
        )}
        <Row label="Ha hiányzik">
          <Select
            value={fallbackMode}
            options={[
              { value: "competition", label: "Verseny WCA-oldala" },
              { value: "hide", label: "Ne legyen QR-kód" },
              { value: "custom", label: "Egyéni tartalék" },
            ]}
            onChange={(mode) =>
              set(
                {
                  fallback:
                    mode === "hide"
                      ? []
                      : mode === "competition"
                        ? [{ variable: "competition.url" }]
                        : [{ text: "https://" }],
                },
                "qr-fallback",
              )
            }
          />
        </Row>
        {fallbackMode === "custom" && (
          <SegmentsField id={`${element.id}-qr-fallback`} kind={kind} segments={element.fallback} onChange={(fallback) => set({ fallback }, "qr-fallback-text")} />
        )}
        <p className="text-xs text-muted-foreground">
          A tartalék akkor kell, ha egy adatmező üres — például az újoncnak a verseny után, a hivatalos eredmények
          felkerüléséig nincs WCA ID-ja.
        </p>
        {preview !== null && (
          <p className="break-all text-xs text-muted-foreground">
            Előnézetben kódolva: <span className="tabular font-mono text-foreground">{preview}</span>
          </p>
        )}
      </Section>

      <Section title="QR-kód megjelenése">
        <Row label="Hibajavítás">
          <Select
            value={element.errorCorrection}
            options={[
              { value: "L", label: "Alacsony (7%)" },
              { value: "M", label: "Közepes (15%)" },
              { value: "Q", label: "Magas (25%)" },
              { value: "H", label: "Legmagasabb (30%)" },
            ]}
            onChange={(errorCorrection) => set({ errorCorrection }, "qr-ecc")}
          />
        </Row>
        <Row label="Szín">
          <ColorInput value={element.color} allowMedal={kind === "placement"} onChange={(color) => set({ color: color ?? "#16181d" }, "qr-color")} />
        </Row>
        <Row label="Háttér">
          <ColorInput value={element.background} allowNone onChange={(background) => set({ background }, "qr-background")} />
        </Row>
        <Row label="Csendzóna">
          <NumberInput value={element.quietZone} min={0} max={4} step={1} decimals={0} unit="pont" onChange={(quietZone) => set({ quietZone }, "qr-quiet")} />
        </Row>
        <p className="text-xs text-muted-foreground">
          Nyomtatva legalább 20 mm-es méret ajánlott. A kód mindig négyzet, a dobozban középre kerül, és vektorosan
          kerül a PDF-be.
        </p>
      </Section>
    </>
  );
}

function ElementProperties(props: Props & { element: CertificateElement }) {
  const { element, kind, onUpdate } = props;
  const set = (patch: Partial<CertificateElement>, key: string) =>
    onUpdate((e) => ({ ...e, ...patch }) as CertificateElement, `${element.id}:${key}`);
  const asset = element.type === "image" ? props.assets.find((a) => a.id === element.assetId) : undefined;
  const dpi = asset && element.type === "image" ? Math.round(asset.width / (element.width / 25.4)) : null;

  return (
    <>
      <Section title={ELEMENT_LABELS[element.type]}>
        <Row label="Név" htmlFor="element-name">
          <TextInput id="element-name" value={element.name} maxLength={80} onChange={(name) => set({ name }, "name")} />
        </Row>
        <div className="grid grid-cols-2 gap-2">
          <NumberInput value={element.x} step={1} unit="X mm" ariaLabel="Vízszintes pozíció" onChange={(x) => set({ x }, "x")} />
          <NumberInput value={element.y} step={1} unit="Y mm" ariaLabel="Függőleges pozíció" onChange={(y) => set({ y }, "y")} />
          <NumberInput value={element.width} min={2} step={1} unit="Sz mm" ariaLabel="Szélesség" onChange={(width) => set({ width }, "width")} />
          <NumberInput value={element.height} min={2} step={1} unit="M mm" ariaLabel="Magasság" onChange={(height) => set({ height }, "height")} />
        </div>
        <Row label="Átlátszatlanság">
          <NumberInput value={Math.round(element.opacity * 100)} min={0} max={100} step={5} decimals={0} unit="%" onChange={(v) => set({ opacity: v / 100 }, "opacity")} />
        </Row>
        <div className="flex flex-wrap gap-1">
          <ToggleGroup
            label="Rétegsorrend"
            value=""
            options={[
              { value: "front", label: <MdFlipToFront className="h-4 w-4" />, title: "Előtérbe" },
              { value: "forward", label: "↑", title: "Eggyel előrébb" },
              { value: "backward", label: "↓", title: "Eggyel hátrébb" },
              { value: "back", label: <MdFlipToBack className="h-4 w-4" />, title: "Háttérbe" },
            ]}
            onChange={(move) => props.onLayer(move as LayerMove)}
          />
          <ToggleGroup
            label="Elem műveletek"
            value={[element.locked ? "lock" : "", element.hidden ? "hide" : ""].filter(Boolean) as string[]}
            options={[
              { value: "lock", label: element.locked ? <FaLock className="h-3 w-3" /> : <FaLockOpen className="h-3 w-3" />, title: element.locked ? "Zárolás feloldása" : "Zárolás" },
              { value: "hide", label: element.hidden ? <FaRegEyeSlash className="h-3 w-3" /> : <FaRegEye className="h-3 w-3" />, title: element.hidden ? "Megjelenítés" : "Elrejtés" },
              { value: "copy", label: <FaRegClone className="h-3 w-3" />, title: "Duplikálás (Ctrl+D)" },
              { value: "delete", label: <FaRegTrashAlt className="h-3 w-3" />, title: "Törlés (Del)" },
            ]}
            onChange={(action) => {
              if (action === "lock") set({ locked: !element.locked }, "locked");
              if (action === "hide") set({ hidden: !element.hidden }, "hidden");
              if (action === "copy") props.onDuplicate();
              if (action === "delete") props.onDelete();
            }}
          />
        </div>
      </Section>

      {element.type === "text" && (
        <TextProperties element={element} kind={kind} overflow={props.overflow} onUpdate={onUpdate} contentEditorRef={props.contentEditorRef} />
      )}

      {element.type === "image" && (
        <Section title="Kép">
          <Row label="Illesztés">
            <Select
              value={element.fit}
              options={[
                { value: "contain", label: "Teljes kép (arányos)" },
                { value: "cover", label: "Kitöltés vágással" },
                { value: "fill", label: "Nyújtás" },
              ]}
              onChange={(fit) => set({ fit }, "fit")}
            />
          </Row>
          {dpi !== null && (
            <p className={`text-xs ${dpi < 150 ? "text-brand-red" : "text-muted-foreground"}`}>
              Nyomtatási felbontás ebben a méretben: <span className="tabular font-mono">{dpi}</span> dpi
              {dpi < 150 ? " — kevés, a nyomat pixeles lehet." : dpi >= 300 ? " — kiváló." : " — megfelelő."}
            </p>
          )}
          <AssetLibrary
            assets={props.assets}
            assetUrls={props.assetUrls}
            selectedId={element.assetId}
            onPick={(picked) => set({ assetId: picked.id }, "asset")}
            onUpload={props.onUpload}
          />
        </Section>
      )}

      {(element.type === "rect" || element.type === "ellipse") && (
        <Section title="Kitöltés és körvonal">
          <Row label="Kitöltés">
            <ColorInput value={element.fill} allowNone allowMedal={kind === "placement"} onChange={(fill) => set({ fill }, "fill")} />
          </Row>
          <Row label="Körvonal">
            <ColorInput value={element.stroke} allowNone allowMedal={kind === "placement"} onChange={(stroke) => set({ stroke }, "stroke")} />
          </Row>
          <Row label="Vastagság">
            <NumberInput value={element.strokeWidth} min={0.1} max={20} step={0.25} decimals={2} unit="pt" onChange={(strokeWidth) => set({ strokeWidth }, "strokeWidth")} />
          </Row>
          {element.type === "rect" && (
            <Row label="Lekerekítés">
              <NumberInput value={element.radius} min={0} max={100} step={1} unit="mm" onChange={(radius) => set({ radius }, "radius")} />
            </Row>
          )}
        </Section>
      )}

      {element.type === "line" && (
        <Section title="Vonal">
          <Row label="Irány">
            <Select
              value={element.orientation}
              options={[
                { value: "horizontal", label: "Vízszintes" },
                { value: "vertical", label: "Függőleges" },
              ]}
              onChange={(orientation) => set({ orientation }, "orientation")}
            />
          </Row>
          <Row label="Szín">
            <ColorInput value={element.stroke} allowMedal={kind === "placement"} onChange={(stroke) => set({ stroke: stroke ?? "#16181d" }, "stroke")} />
          </Row>
          <Row label="Vastagság">
            <NumberInput value={element.strokeWidth} min={0.1} max={20} step={0.25} decimals={2} unit="pt" onChange={(strokeWidth) => set({ strokeWidth }, "strokeWidth")} />
          </Row>
        </Section>
      )}

      {element.type === "qr" && <QrProperties element={element} kind={kind} preview={props.qrPreview} set={set} />}

      {element.type === "logo" && (
        <Section title="Logó">
          <Row label="Feliratszín">
            <ColorInput value={element.color} allowMedal={kind === "placement"} onChange={(color) => set({ color: color ?? "#16181d" }, "color")} />
          </Row>
          <p className="text-xs text-muted-foreground">A logó vektorosan kerül a PDF-be, bármekkora méretben éles. A piros és zöld kocka színe rögzített.</p>
        </Section>
      )}
    </>
  );
}

function PageProperties(props: Props) {
  const { page } = props.document;
  const [scaleElements, setScaleElements] = useState(true);
  const preset = PAGE_PRESETS.find((p) => Math.abs(p.width - page.width) < 0.5 && Math.abs(p.height - page.height) < 0.5);

  return (
    <>
      <Section title="Sablon">
        <Row label="Név" htmlFor="template-name">
          <TextInput id="template-name" value={props.name} maxLength={120} onChange={props.onRename} />
        </Row>
        <Row label="Leírás" htmlFor="template-description">
          <textarea
            id="template-description"
            value={props.description}
            maxLength={500}
            rows={2}
            onChange={(e) => props.onDescribe(e.target.value)}
            className="w-full rounded-control border border-line bg-surface px-2.5 py-1.5 text-sm text-foreground focus:border-foreground focus:outline-none"
          />
        </Row>
        <p className="text-xs text-muted-foreground">
          {KIND_LABELS[props.kind].title} — {KIND_LABELS[props.kind].description}
        </p>
      </Section>

      <Section title="Oldal">
        <Row label="Méret">
          <Select
            value={preset?.id ?? "custom"}
            options={[...PAGE_PRESETS.map((p) => ({ value: p.id as string, label: p.label })), ...(preset ? [] : [{ value: "custom", label: "Egyéni" }])]}
            onChange={(id) => {
              const next = PAGE_PRESETS.find((p) => p.id === id);
              if (next) props.onPage({ ...page, width: next.width, height: next.height }, scaleElements);
            }}
          />
        </Row>
        <div className="grid grid-cols-2 gap-2">
          <NumberInput value={page.width} min={50} max={500} step={1} unit="Sz mm" ariaLabel="Oldalszélesség" onChange={(width) => props.onPage({ ...page, width }, scaleElements)} />
          <NumberInput value={page.height} min={50} max={500} step={1} unit="M mm" ariaLabel="Oldalmagasság" onChange={(height) => props.onPage({ ...page, height }, scaleElements)} />
        </div>
        <Checkbox checked={scaleElements} onChange={setScaleElements}>
          Méretváltáskor az elemek is arányosan változzanak
        </Checkbox>
        <Row label="Háttérszín">
          <ColorInput value={page.background} onChange={(background) => props.onPage({ ...page, background: background ?? "#ffffff" }, false)} />
        </Row>
        <Row label="Biztonsági margó">
          <NumberInput value={page.margin} min={0} max={40} step={1} unit="mm" onChange={(margin) => props.onPage({ ...page, margin }, false)} />
        </Row>
        <p className="text-xs text-muted-foreground">
          A szaggatott keret a margó: a legtöbb irodai nyomtató 4–6 mm-re a lap szélétől már nem nyomtat.
        </p>
      </Section>

      <Section title="Háttérkép">
        <p className="text-xs text-muted-foreground">Válassz egy képet: a teljes lapot kitölti, zárolva, minden más mögé kerül.</p>
        <AssetLibrary assets={props.assets} assetUrls={props.assetUrls} onPick={props.onBackground} onUpload={props.onUpload} />
      </Section>
    </>
  );
}

export default function PropertiesPanel(props: Props) {
  return (
    <div className="flex flex-col">
      {props.element ? <ElementProperties {...props} element={props.element} /> : <PageProperties {...props} />}
    </div>
  );
}
