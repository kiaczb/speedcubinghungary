import { LOGO_ASPECT } from "./model";
import type { ImageFit, LineElement, LogoElement } from "./model";

export type Box = { x: number; y: number; width: number; height: number };

export function logoBox(element: LogoElement): Box {
  const width = Math.min(element.width, element.height * LOGO_ASPECT);
  const height = width / LOGO_ASPECT;
  return {
    x: element.x + (element.width - width) / 2,
    y: element.y + (element.height - height) / 2,
    width,
    height,
  };
}

export function imageBox(box: Box, fit: ImageFit, imageWidth: number, imageHeight: number): Box {
  if (fit === "fill" || imageWidth <= 0 || imageHeight <= 0) return box;
  const scale =
    fit === "contain"
      ? Math.min(box.width / imageWidth, box.height / imageHeight)
      : Math.max(box.width / imageWidth, box.height / imageHeight);
  const width = imageWidth * scale;
  const height = imageHeight * scale;
  return {
    x: box.x + (box.width - width) / 2,
    y: box.y + (box.height - height) / 2,
    width,
    height,
  };
}

export function lineEnds(element: LineElement): [{ x: number; y: number }, { x: number; y: number }] {
  if (element.orientation === "vertical") {
    const x = element.x + element.width / 2;
    return [{ x, y: element.y }, { x, y: element.y + element.height }];
  }
  const y = element.y + element.height / 2;
  return [{ x: element.x, y }, { x: element.x + element.width, y }];
}

export function roundedRectPath(width: number, height: number, radius: number): string {
  const r = Math.max(0, Math.min(radius, width / 2, height / 2));
  if (r === 0) return `M0 0 H${width} V${height} H0 Z`;
  const k = r * 0.5523;
  return [
    `M${r} 0`,
    `H${width - r}`,
    `C${width - r + k} 0 ${width} ${r - k} ${width} ${r}`,
    `V${height - r}`,
    `C${width} ${height - r + k} ${width - r + k} ${height} ${width - r} ${height}`,
    `H${r}`,
    `C${r - k} ${height} 0 ${height - r + k} 0 ${height - r}`,
    `V${r}`,
    `C0 ${r - k} ${r - k} 0 ${r} 0`,
    "Z",
  ].join(" ");
}
