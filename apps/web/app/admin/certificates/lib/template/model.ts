import type { FontFamilyId } from "./fonts";
import type { VariableKey } from "./variables";

export const DOCUMENT_SCHEMA = 1;

export const PT_PER_MM = 72 / 25.4;

export const MEDAL = "medal";

export type ColorValue = string;

export type TemplateKind = "participation" | "placement";

export type TextSegment = { text: string } | { variable: VariableKey };

export type HorizontalAlign = "left" | "center" | "right";

export type VerticalAlign = "top" | "middle" | "bottom";

export type TextFit = "wrap" | "shrink" | "wrapShrink";

export type TextStyle = {
  font: FontFamilyId;
  size: number;
  bold: boolean;
  italic: boolean;
  underline: boolean;
  color: ColorValue;
  align: HorizontalAlign;
  verticalAlign: VerticalAlign;
  lineHeight: number;
  letterSpacing: number;
  uppercase: boolean;
  fit: TextFit;
  minSize: number;
};

type ElementBase = {
  id: string;
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  opacity: number;
  locked?: boolean;
  hidden?: boolean;
};

export type TextElement = ElementBase & {
  type: "text";
  content: TextSegment[];
  style: TextStyle;
};

export type ImageFit = "contain" | "cover" | "fill";

export type ImageElement = ElementBase & {
  type: "image";
  assetId: string | null;
  fit: ImageFit;
};

export type ShapeElement = ElementBase & {
  type: "rect" | "ellipse";
  fill: ColorValue | null;
  stroke: ColorValue | null;
  strokeWidth: number;
  radius: number;
};

export type LineElement = ElementBase & {
  type: "line";
  orientation: "horizontal" | "vertical";
  stroke: ColorValue;
  strokeWidth: number;
};

export type LogoElement = ElementBase & {
  type: "logo";
  color: ColorValue;
};

export type QrErrorCorrection = "L" | "M" | "Q" | "H";

export type QrElement = ElementBase & {
  type: "qr";
  content: TextSegment[];
  fallback: TextSegment[];
  errorCorrection: QrErrorCorrection;
  color: ColorValue;
  background: string | null;
  quietZone: number;
};

export type CertificateElement =
  | TextElement
  | ImageElement
  | ShapeElement
  | LineElement
  | LogoElement
  | QrElement;

export type ElementType = CertificateElement["type"];

export type PageSettings = {
  width: number;
  height: number;
  background: string;
  margin: number;
};

export type CertificateDocument = {
  schemaVersion: typeof DOCUMENT_SCHEMA;
  page: PageSettings;
  elements: CertificateElement[];
};

export type CertificateTemplate = {
  id: string;
  name: string;
  description: string;
  kind: TemplateKind;
  document: CertificateDocument;
  version: number;
  updatedByName: string;
  createdAt: string;
  updatedAt: string;
};

export function blankDocument(): CertificateDocument {
  return {
    schemaVersion: DOCUMENT_SCHEMA,
    page: { width: 297, height: 210, background: "#ffffff", margin: 10 },
    elements: [],
  };
}

export const LOGO_ASPECT = 1375 / 300;

export const PAGE_PRESETS = [
  { id: "a4-landscape", label: "A4 fekvő", width: 297, height: 210 },
  { id: "a4-portrait", label: "A4 álló", width: 210, height: 297 },
  { id: "a5-landscape", label: "A5 fekvő", width: 210, height: 148 },
  { id: "a5-portrait", label: "A5 álló", width: 148, height: 210 },
  { id: "a3-landscape", label: "A3 fekvő", width: 420, height: 297 },
  { id: "letter-landscape", label: "Letter fekvő", width: 279.4, height: 215.9 },
] as const;

export const MEDAL_COLORS: Record<number, string> = {
  1: "#a8832a",
  2: "#7f858f",
  3: "#9c5f3a",
};

export const INK = "#16181d";

export function resolveColor(color: ColorValue, place?: number): string {
  if (color !== MEDAL) return color;
  return (place && MEDAL_COLORS[place]) || INK;
}

export function elementSegments(element: CertificateElement): TextSegment[] {
  if (element.type === "text") return element.content;
  if (element.type === "qr") return [...element.content, ...element.fallback];
  return [];
}

export type CertificateAsset = {
  id: string;
  name: string;
  mime: string;
  width: number;
  height: number;
  size: number;
  createdAt: string;
};
