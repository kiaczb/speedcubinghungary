import { INK } from "../../../lib/template/model";
import type {
  CertificateElement,
  ElementType,
  PageSettings,
  TextSegment,
  TextStyle,
} from "../../../lib/template/model";
import { variableLabel } from "../../../lib/template/variables";
import type { VariableKey } from "../../../lib/template/variables";

export const MIN_SIZE = 2;

export const DRAG_TYPE = "application/x-certificate-element";

export const ELEMENT_LABELS: Record<ElementType, string> = {
  text: "Szöveg",
  image: "Kép",
  logo: "Logó",
  rect: "Téglalap",
  ellipse: "Ellipszis",
  line: "Vonal",
  qr: "QR-kód",
};

export const QR_PRESETS: { id: string; label: string; content: TextSegment[]; fallback: TextSegment[] }[] = [
  {
    id: "profile",
    label: "Versenyző WCA-profilja",
    content: [{ text: "https://www.worldcubeassociation.org/persons/" }, { variable: "competitor.wcaId" }],
    fallback: [{ variable: "competition.url" }],
  },
  {
    id: "competition",
    label: "Verseny WCA-oldala",
    content: [{ variable: "competition.url" }],
    fallback: [],
  },
];

export const DEFAULT_TEXT_STYLE: TextStyle = {
  font: "ebgaramond",
  size: 18,
  bold: false,
  italic: false,
  underline: false,
  color: INK,
  align: "center",
  verticalAlign: "middle",
  lineHeight: 1.2,
  letterSpacing: 0,
  uppercase: false,
  fit: "wrap",
  minSize: 8,
};

export function newId(): string {
  return crypto.randomUUID().slice(0, 8);
}

export type NewElement =
  | { type: Exclude<ElementType, "image" | "text"> }
  | { type: "text"; variable?: VariableKey }
  | { type: "image"; assetId: string | null; aspect?: number };

function centered(width: number, height: number, at: { x: number; y: number }, page: PageSettings) {
  const x = Math.min(Math.max(at.x - width / 2, 0), Math.max(page.width - width, 0));
  const y = Math.min(Math.max(at.y - height / 2, 0), Math.max(page.height - height, 0));
  return { x: Math.round(x * 10) / 10, y: Math.round(y * 10) / 10, width, height };
}

export function createElement(spec: NewElement, at: { x: number; y: number }, page: PageSettings): CertificateElement {
  const base = { id: newId(), opacity: 1 };
  switch (spec.type) {
    case "text": {
      const isName = spec.variable === "competitor.name" || spec.variable === "competitor.fullName";
      const box = centered(isName ? 200 : 110, isName ? 20 : 12, at, page);
      return {
        ...base,
        ...box,
        type: "text",
        name: spec.variable ? variableLabel(spec.variable) : ELEMENT_LABELS.text,
        content: spec.variable ? [{ variable: spec.variable }] : [{ text: "Szöveg" }],
        style: isName
          ? { ...DEFAULT_TEXT_STYLE, size: 40, bold: true, fit: "shrink", minSize: 20 }
          : { ...DEFAULT_TEXT_STYLE },
      };
    }
    case "image": {
      const width = 50;
      const height = spec.aspect ? Math.round((width / spec.aspect) * 10) / 10 : 35;
      return { ...base, ...centered(width, height, at, page), type: "image", name: ELEMENT_LABELS.image, assetId: spec.assetId, fit: "contain" };
    }
    case "logo":
      return { ...base, ...centered(66, 14.4, at, page), type: "logo", name: ELEMENT_LABELS.logo, color: INK };
    case "rect":
      return { ...base, ...centered(60, 30, at, page), type: "rect", name: ELEMENT_LABELS.rect, fill: null, stroke: INK, strokeWidth: 1, radius: 0 };
    case "ellipse":
      return { ...base, ...centered(30, 30, at, page), type: "ellipse", name: ELEMENT_LABELS.ellipse, fill: null, stroke: INK, strokeWidth: 1, radius: 0 };
    case "line":
      return { ...base, ...centered(60, 2, at, page), type: "line", name: ELEMENT_LABELS.line, orientation: "horizontal", stroke: INK, strokeWidth: 1 };
    case "qr": {
      const preset = QR_PRESETS[0];
      return {
        ...base,
        ...centered(25, 25, at, page),
        type: "qr",
        name: ELEMENT_LABELS.qr,
        content: preset.content,
        fallback: preset.fallback,
        errorCorrection: "M",
        color: INK,
        background: null,
        quietZone: 2,
      };
    }
  }
}

export function duplicate(element: CertificateElement, offset = 5): CertificateElement {
  return {
    ...structuredClone(element),
    id: newId(),
    name: `${element.name} (másolat)`,
    x: element.x + offset,
    y: element.y + offset,
    locked: false,
  };
}

export function backgroundImage(assetId: string, page: PageSettings): CertificateElement {
  return {
    id: newId(),
    name: "Háttérkép",
    type: "image",
    assetId,
    fit: "cover",
    x: 0,
    y: 0,
    width: page.width,
    height: page.height,
    opacity: 1,
    locked: true,
  };
}
