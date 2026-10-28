import { PT_PER_MM } from "./model";
import type { TextElement } from "./model";
import { fontFile } from "./fonts";

export type Measurer = {
  width(file: string, text: string, size: number): number;
  ascent(file: string, size: number): number;
  descent(file: string, size: number): number;
  covers(file: string, text: string): boolean;
};

export type LaidOutLine = {
  text: string;
  x: number;
  baseline: number;
  width: number;
};

export type TextLayout = {
  file: string;
  size: number;
  letterSpacing: number;
  lines: LaidOutLine[];
  overflow: boolean;
  covered: boolean;
};

const SIZE_STEP = 0.25;
const TOLERANCE = 0.01;

function wrapParagraph(paragraph: string, fits: (line: string) => boolean): string[] {
  const words = paragraph.split(/ +/).filter(Boolean);
  if (words.length === 0) return [""];
  const lines: string[] = [];
  let line = "";
  for (const word of words) {
    const candidate = line ? `${line} ${word}` : word;
    if (line && !fits(candidate)) {
      lines.push(line);
      line = word;
    } else {
      line = candidate;
    }
  }
  lines.push(line);
  return lines;
}

export function layoutText(element: TextElement, rawText: string, measurer: Measurer): TextLayout {
  const style = element.style;
  const text = style.uppercase ? rawText.toLocaleUpperCase("hu-HU") : rawText;
  const file = fontFile(style.font, style.bold, style.italic);
  const boxWidth = element.width * PT_PER_MM;
  const boxHeight = element.height * PT_PER_MM;
  const paragraphs = text.split("\n");

  const widthOf = (line: string, size: number) =>
    measurer.width(file, line, size) + style.letterSpacing * Math.max([...line].length - 1, 0);
  const linesAt = (size: number) =>
    style.fit === "shrink"
      ? paragraphs
      : paragraphs.flatMap((p) => wrapParagraph(p, (line) => widthOf(line, size) <= boxWidth + TOLERANCE));
  const fits = (lines: string[], size: number) =>
    lines.length * size * style.lineHeight <= boxHeight + TOLERANCE &&
    lines.every((line) => widthOf(line, size) <= boxWidth + TOLERANCE);

  let size = style.size;
  let lines = linesAt(size);
  let overflow = !fits(lines, size);
  if (overflow && style.fit !== "wrap") {
    const floor = Math.min(style.minSize, style.size);
    for (let candidate = style.size - SIZE_STEP; candidate >= floor - TOLERANCE; candidate -= SIZE_STEP) {
      size = candidate;
      lines = linesAt(size);
      if (fits(lines, size)) {
        overflow = false;
        break;
      }
    }
  }

  const lineBox = size * style.lineHeight;
  const ascent = measurer.ascent(file, size);
  const descent = measurer.descent(file, size);
  const blockHeight = lines.length * lineBox;
  const offset =
    style.verticalAlign === "top"
      ? 0
      : style.verticalAlign === "bottom"
        ? boxHeight - blockHeight
        : (boxHeight - blockHeight) / 2;

  const laidOut = lines.map((line, i) => {
    const width = widthOf(line, size);
    const x =
      style.align === "left" ? 0 : style.align === "right" ? boxWidth - width : (boxWidth - width) / 2;
    const baseline = offset + i * lineBox + (lineBox - (ascent + descent)) / 2 + ascent;
    return { text: line, x: x / PT_PER_MM, baseline: baseline / PT_PER_MM, width: width / PT_PER_MM };
  });

  return {
    file,
    size,
    letterSpacing: style.letterSpacing,
    lines: laidOut,
    overflow,
    covered: measurer.covers(file, text.replace(/\s/g, "")),
  };
}
