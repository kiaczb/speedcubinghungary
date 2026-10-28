import type { CertificateDocument } from "./model";

export type FontFamilyId = "ebgaramond" | "playfair" | "geist" | "greatvibes";

type FontFiles = {
  regular: string;
  bold?: string;
  italic?: string;
  boldItalic?: string;
};

export const FONT_FAMILIES: { id: FontFamilyId; label: string; files: FontFiles }[] = [
  {
    id: "ebgaramond",
    label: "EB Garamond (talpas)",
    files: {
      regular: "EBGaramond-Regular.ttf",
      bold: "EBGaramond-SemiBold.ttf",
      italic: "EBGaramond-Italic.ttf",
      boldItalic: "EBGaramond-SemiBoldItalic.ttf",
    },
  },
  {
    id: "playfair",
    label: "Playfair Display (díszes talpas)",
    files: { regular: "PlayfairDisplay-Regular.ttf", bold: "PlayfairDisplay-Bold.ttf" },
  },
  {
    id: "geist",
    label: "Geist (talpatlan)",
    files: { regular: "Geist-Regular.ttf", bold: "Geist-SemiBold.ttf" },
  },
  {
    id: "greatvibes",
    label: "Great Vibes (kézírásos)",
    files: { regular: "GreatVibes-Regular.ttf" },
  },
];

export const FONT_BASE_URL = "/fonts/certificate/";

export function fontFamily(id: FontFamilyId) {
  return FONT_FAMILIES.find((f) => f.id === id) ?? FONT_FAMILIES[0];
}

export function fontFile(id: FontFamilyId, bold: boolean, italic: boolean): string {
  const { files } = fontFamily(id);
  if (bold && italic && files.boldItalic) return files.boldItalic;
  if (bold && files.bold) return files.bold;
  if (italic && files.italic) return files.italic;
  return files.regular;
}

export function supportsBold(id: FontFamilyId): boolean {
  return Boolean(fontFamily(id).files.bold);
}

export function supportsItalic(id: FontFamilyId): boolean {
  return Boolean(fontFamily(id).files.italic);
}

export function cssFontFamily(file: string): string {
  return `cert-${file.replace(/\.ttf$/, "")}`;
}

export function documentFontFiles(document: CertificateDocument): string[] {
  const files = new Set<string>();
  for (const element of document.elements) {
    if (element.type === "text") {
      files.add(fontFile(element.style.font, element.style.bold, element.style.italic));
    }
  }
  return [...files];
}
