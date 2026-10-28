import { PDFDocument } from "pdf-lib";
import type { PDFFont } from "pdf-lib";
import fontkit from "@pdf-lib/fontkit";
import type { Measurer } from "./textLayout";

export type FontBytes = Record<string, ArrayBuffer>;

export async function embedFonts(pdf: PDFDocument, bytes: FontBytes): Promise<Record<string, PDFFont>> {
  pdf.registerFontkit(fontkit);
  const fonts: Record<string, PDFFont> = {};
  for (const [file, data] of Object.entries(bytes)) {
    fonts[file] = await pdf.embedFont(data);
  }
  return fonts;
}

export function measurerFor(fonts: Record<string, PDFFont>): Measurer {
  const charsets = new Map<string, Set<number>>();
  const font = (file: string) => {
    const found = fonts[file];
    if (!found) throw new Error(`Nincs betöltve: ${file}`);
    return found;
  };
  return {
    width: (file, text, size) => font(file).widthOfTextAtSize(text, size),
    ascent: (file, size) => font(file).heightAtSize(size, { descender: false }),
    descent: (file, size) =>
      font(file).heightAtSize(size) - font(file).heightAtSize(size, { descender: false }),
    covers: (file, text) => {
      let charset = charsets.get(file);
      if (!charset) {
        charset = new Set(font(file).getCharacterSet());
        charsets.set(file, charset);
      }
      const available = charset;
      return [...text].every((c) => available.has(c.codePointAt(0) ?? 0));
    },
  };
}

export async function createMeasurer(bytes: FontBytes): Promise<Measurer> {
  return measurerFor(await embedFonts(await PDFDocument.create(), bytes));
}
