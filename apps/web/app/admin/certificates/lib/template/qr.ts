import qrcode from "qrcode-generator";
import type { QrElement } from "./model";
import { hasMissingValue, segmentsText } from "./variables";
import type { CertificateContext } from "./variables";
import type { Box } from "./geometry";

export const MIN_MODULE_MM = 0.4;

export type QrMatrix = {
  count: number;
  path: string;
};

export type QrLayout =
  | { state: "hidden" }
  | { state: "error"; text: string }
  | { state: "ok"; text: string; matrix: QrMatrix; box: Box; module: number; tooSmall: boolean };

const cache = new Map<string, QrMatrix | null>();

function binary(text: string): string {
  return Array.from(new TextEncoder().encode(text), (byte) => String.fromCharCode(byte)).join("");
}

export function qrMatrix(text: string, errorCorrection: QrElement["errorCorrection"]): QrMatrix | null {
  const key = `${errorCorrection}|${text}`;
  if (cache.has(key)) return cache.get(key) ?? null;

  let matrix: QrMatrix | null = null;
  try {
    const qr = qrcode(0, errorCorrection);
    qr.addData(binary(text), "Byte");
    qr.make();
    const count = qr.getModuleCount();
    const runs: string[] = [];
    for (let row = 0; row < count; row++) {
      let col = 0;
      while (col < count) {
        if (!qr.isDark(row, col)) {
          col++;
          continue;
        }
        const start = col;
        while (col < count && qr.isDark(row, col)) col++;
        runs.push(`M${start} ${row}h${col - start}v1h${start - col}z`);
      }
    }
    matrix = { count, path: runs.join("") };
  } catch {
    matrix = null;
  }
  if (cache.size > 2000) cache.clear();
  cache.set(key, matrix);
  return matrix;
}

export function qrText(element: QrElement, context: CertificateContext | null): string | null {
  if (!context) return segmentsText(element.content, null);
  const segments = hasMissingValue(element.content, context) ? element.fallback : element.content;
  if (segments.length === 0 || hasMissingValue(segments, context)) return null;
  const text = segmentsText(segments, context).trim();
  return text || null;
}

export function layoutQr(element: QrElement, context: CertificateContext | null): QrLayout {
  const text = qrText(element, context);
  if (text === null) return { state: "hidden" };
  const matrix = qrMatrix(text, element.errorCorrection);
  if (!matrix) return { state: "error", text };

  const side = Math.min(element.width, element.height);
  const moduleSize = side / (matrix.count + 2 * element.quietZone);
  return {
    state: "ok",
    text,
    matrix,
    module: moduleSize,
    tooSmall: moduleSize < MIN_MODULE_MM,
    box: {
      x: element.x + (element.width - side) / 2,
      y: element.y + (element.height - side) / 2,
      width: side,
      height: side,
    },
  };
}
