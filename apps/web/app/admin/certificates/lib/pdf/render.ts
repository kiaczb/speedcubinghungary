import {
  PDFDocument,
  clip,
  closePath,
  concatTransformationMatrix,
  endPath,
  lineTo,
  moveTo,
  popGraphicsState,
  pushGraphicsState,
  rgb,
  setCharacterSpacing,
} from "pdf-lib";
import type { PDFEmbeddedPage, PDFFont, PDFImage, PDFPage, RGB } from "pdf-lib";
import {
  LOGO_BOTTOM_COLOR,
  LOGO_BOTTOM_PATH,
  LOGO_CUBE_PATH,
  LOGO_FRAME_PATHS,
  LOGO_LETTER_PATHS,
  LOGO_TOP_COLOR,
  LOGO_TOP_PATH,
} from "../../../../components/svgs/speedcubingLogo";
import { PT_PER_MM, resolveColor } from "../template/model";
import type {
  CertificateDocument,
  CertificateElement,
  ImageElement,
  LineElement,
  LogoElement,
  QrElement,
  ShapeElement,
  TextElement,
} from "../template/model";
import { layoutQr } from "../template/qr";
import { imageBox, lineEnds, logoBox, roundedRectPath } from "../template/geometry";
import type { Box } from "../template/geometry";
import { layoutText } from "../template/textLayout";
import type { Measurer } from "../template/textLayout";
import { embedFonts, measurerFor } from "../template/metrics";
import type { FontBytes } from "../template/metrics";
import { segmentsText } from "../template/variables";
import type { CertificateContext } from "../template/variables";

export type AssetBytes = Record<string, { bytes: ArrayBuffer; mime: string }>;

export type CertificateMeta = {
  title: string;
  author: string;
};

export type RenderedCertificates = {
  bytes: Uint8Array;
  overflowing: string[];
  unsupported: string[];
  qrFailed: string[];
  missingImages: number;
};

const LOGO_UNITS = { minX: 800, minY: 1300, width: 13750, height: 3000 };

function hex(value: string): RGB {
  const n = Number.parseInt(value.replace("#", "").slice(0, 6), 16) || 0;
  return rgb(((n >> 16) & 0xff) / 255, ((n >> 8) & 0xff) / 255, (n & 0xff) / 255);
}

type Canvas = {
  page: PDFPage;
  height: number;
  place?: number;
};

const pt = (mm: number) => mm * PT_PER_MM;

function pdfY(canvas: Canvas, topMm: number): number {
  return canvas.height - pt(topMm);
}

async function embedLogo(pdf: PDFDocument, color: string): Promise<PDFEmbeddedPage> {
  const source = await PDFDocument.create();
  const page = source.addPage([LOGO_UNITS.width, LOGO_UNITS.height]);
  const ink = hex(color);
  page.pushOperators(pushGraphicsState(), concatTransformationMatrix(1, 0, 0, -1, 0, 0));
  const at = { x: -LOGO_UNITS.minX, y: LOGO_UNITS.minY, scale: 1 };
  for (const d of [...LOGO_LETTER_PATHS, ...LOGO_FRAME_PATHS, LOGO_CUBE_PATH]) {
    page.drawSvgPath(d, { ...at, color: ink });
  }
  page.drawSvgPath(LOGO_TOP_PATH, { ...at, color: hex(LOGO_TOP_COLOR) });
  page.drawSvgPath(LOGO_BOTTOM_PATH, { ...at, color: hex(LOGO_BOTTOM_COLOR) });
  page.pushOperators(popGraphicsState());
  return pdf.embedPage(page);
}

function drawText(
  canvas: Canvas,
  element: TextElement,
  context: CertificateContext,
  fonts: Record<string, PDFFont>,
  measurer: Measurer,
) {
  const layout = layoutText(element, segmentsText(element.content, context), measurer);
  const font = fonts[layout.file];
  const color = hex(resolveColor(element.style.color, canvas.place));
  if (layout.letterSpacing) canvas.page.pushOperators(setCharacterSpacing(layout.letterSpacing));
  for (const line of layout.lines) {
    if (!line.text) continue;
    const x = pt(element.x + line.x);
    const y = pdfY(canvas, element.y + line.baseline);
    canvas.page.drawText(line.text, { x, y, font, size: layout.size, color, opacity: element.opacity });
    if (element.style.underline) {
      const offset = layout.size * 0.12;
      canvas.page.drawLine({
        start: { x, y: y - offset },
        end: { x: x + pt(line.width), y: y - offset },
        thickness: Math.max(layout.size * 0.06, 0.4),
        color,
        opacity: element.opacity,
      });
    }
  }
  if (layout.letterSpacing) canvas.page.pushOperators(setCharacterSpacing(0));
  return layout;
}

function drawImage(canvas: Canvas, element: ImageElement, image: PDFImage) {
  const target = imageBox(element, element.fit, image.width, image.height);
  const clipped = element.fit === "cover";
  if (clipped) {
    const left = pt(element.x);
    const right = pt(element.x + element.width);
    const top = pdfY(canvas, element.y);
    const bottom = pdfY(canvas, element.y + element.height);
    canvas.page.pushOperators(
      pushGraphicsState(),
      moveTo(left, bottom),
      lineTo(right, bottom),
      lineTo(right, top),
      lineTo(left, top),
      closePath(),
      clip(),
      endPath(),
    );
  }
  canvas.page.drawImage(image, {
    x: pt(target.x),
    y: pdfY(canvas, target.y + target.height),
    width: pt(target.width),
    height: pt(target.height),
    opacity: element.opacity,
  });
  if (clipped) canvas.page.pushOperators(popGraphicsState());
}

function drawShape(canvas: Canvas, element: ShapeElement) {
  const fill = element.fill ? hex(resolveColor(element.fill, canvas.place)) : undefined;
  const stroke = element.stroke ? hex(resolveColor(element.stroke, canvas.place)) : undefined;
  const borderWidth = stroke ? element.strokeWidth : 0;
  if (element.type === "ellipse") {
    canvas.page.drawEllipse({
      x: pt(element.x + element.width / 2),
      y: pdfY(canvas, element.y + element.height / 2),
      xScale: pt(element.width / 2),
      yScale: pt(element.height / 2),
      color: fill,
      borderColor: stroke,
      borderWidth,
      opacity: element.opacity,
      borderOpacity: element.opacity,
    });
    return;
  }
  canvas.page.drawSvgPath(roundedRectPath(pt(element.width), pt(element.height), pt(element.radius)), {
    x: pt(element.x),
    y: pdfY(canvas, element.y),
    color: fill,
    borderColor: stroke,
    borderWidth,
    opacity: element.opacity,
    borderOpacity: element.opacity,
  });
}

function drawLine(canvas: Canvas, element: LineElement) {
  const [start, end] = lineEnds(element);
  canvas.page.drawLine({
    start: { x: pt(start.x), y: pdfY(canvas, start.y) },
    end: { x: pt(end.x), y: pdfY(canvas, end.y) },
    thickness: element.strokeWidth,
    color: hex(resolveColor(element.stroke, canvas.place)),
    opacity: element.opacity,
  });
}

function drawLogo(canvas: Canvas, element: LogoElement, logo: PDFEmbeddedPage) {
  const box: Box = logoBox(element);
  canvas.page.drawPage(logo, {
    x: pt(box.x),
    y: pdfY(canvas, box.y + box.height),
    width: pt(box.width),
    height: pt(box.height),
    opacity: element.opacity,
  });
}

function drawQr(canvas: Canvas, element: QrElement, context: CertificateContext): boolean {
  const layout = layoutQr(element, context);
  if (layout.state === "hidden") return true;
  if (layout.state === "error") return false;
  const { box } = layout;
  if (element.background) {
    canvas.page.drawRectangle({
      x: pt(box.x),
      y: pdfY(canvas, box.y + box.height),
      width: pt(box.width),
      height: pt(box.height),
      color: hex(element.background),
      opacity: element.opacity,
    });
  }
  const modulePt = pt(layout.module);
  canvas.page.drawSvgPath(layout.matrix.path, {
    x: pt(box.x) + element.quietZone * modulePt,
    y: pdfY(canvas, box.y) - element.quietZone * modulePt,
    scale: modulePt,
    color: hex(resolveColor(element.color, canvas.place)),
    opacity: element.opacity,
  });
  return !layout.tooSmall;
}

export async function renderCertificates(
  document: CertificateDocument,
  contexts: CertificateContext[],
  fontBytes: FontBytes,
  assets: AssetBytes,
  meta: CertificateMeta,
): Promise<RenderedCertificates> {
  const pdf = await PDFDocument.create();
  pdf.setTitle(meta.title);
  pdf.setAuthor(meta.author);
  pdf.setCreator(meta.author);
  pdf.setLanguage("hu-HU");

  const fonts = await embedFonts(pdf, fontBytes);
  const measurer = measurerFor(fonts);
  const visible = document.elements.filter((element) => !element.hidden);

  const images = new Map<string, PDFImage>();
  let missingImages = 0;
  for (const element of visible) {
    if (element.type !== "image" || !element.assetId || images.has(element.assetId)) continue;
    const asset = assets[element.assetId];
    if (!asset) {
      missingImages++;
      continue;
    }
    images.set(
      element.assetId,
      asset.mime === "image/png" ? await pdf.embedPng(asset.bytes) : await pdf.embedJpg(asset.bytes),
    );
  }

  const logos = new Map<string, PDFEmbeddedPage>();
  const logoFor = async (element: LogoElement, place?: number) => {
    const color = resolveColor(element.color, place);
    let logo = logos.get(color);
    if (!logo) {
      logo = await embedLogo(pdf, color);
      logos.set(color, logo);
    }
    return logo;
  };

  const overflowing = new Set<string>();
  const unsupported = new Set<string>();
  const qrFailed = new Set<string>();
  const width = pt(document.page.width);
  const height = pt(document.page.height);

  for (const context of contexts) {
    const page = pdf.addPage([width, height]);
    const canvas: Canvas = { page, height, place: context.place };
    const label = context.values["competitor.name"] || context.values["competitor.fullName"] || "";
    if (document.page.background && document.page.background.toLowerCase() !== "#ffffff") {
      page.drawRectangle({ x: 0, y: 0, width, height, color: hex(document.page.background) });
    }

    for (const element of visible as CertificateElement[]) {
      switch (element.type) {
        case "text": {
          const layout = drawText(canvas, element, context, fonts, measurer);
          if (layout.overflow) overflowing.add(label);
          if (!layout.covered) unsupported.add(label);
          break;
        }
        case "image": {
          const image = element.assetId ? images.get(element.assetId) : undefined;
          if (image) drawImage(canvas, element, image);
          break;
        }
        case "rect":
        case "ellipse":
          drawShape(canvas, element);
          break;
        case "line":
          drawLine(canvas, element);
          break;
        case "logo":
          drawLogo(canvas, element, await logoFor(element, context.place));
          break;
        case "qr":
          if (!drawQr(canvas, element, context)) qrFailed.add(label);
          break;
      }
    }
  }

  return {
    bytes: await pdf.save(),
    overflowing: [...overflowing].filter(Boolean),
    unsupported: [...unsupported].filter(Boolean),
    qrFailed: [...qrFailed].filter(Boolean),
    missingImages,
  };
}
