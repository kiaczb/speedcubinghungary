import {
  LOGO_BOTTOM_COLOR,
  LOGO_BOTTOM_PATH,
  LOGO_CUBE_PATH,
  LOGO_FRAME_PATHS,
  LOGO_LETTER_PATHS,
  LOGO_TOP_COLOR,
  LOGO_TOP_PATH,
} from "@/app/components/svgs/speedcubingLogo";
import { PT_PER_MM, resolveColor } from "../lib/template/model";
import type { CertificateElement, QrElement, TextElement } from "../lib/template/model";
import { cssFontFamily } from "../lib/template/fonts";
import { lineEnds, logoBox } from "../lib/template/geometry";
import type { Box } from "../lib/template/geometry";
import { layoutText } from "../lib/template/textLayout";
import type { Measurer, TextLayout } from "../lib/template/textLayout";
import { segmentsText } from "../lib/template/variables";
import { layoutQr } from "../lib/template/qr";
import type { CertificateContext } from "../lib/template/variables";

export const OVERFLOW = "#d81f38";

const PRESERVE: Record<string, string> = {
  contain: "xMidYMid meet",
  cover: "xMidYMid slice",
  fill: "none",
};

function TextShape({ element, layout, place }: { element: TextElement; layout: TextLayout; place?: number }) {
  const color = resolveColor(element.style.color, place);
  const fontSize = layout.size / PT_PER_MM;
  return (
    <g opacity={element.opacity}>
      {layout.lines.map((line, i) => (
        <g key={i}>
          <text
            x={element.x + line.x}
            y={element.y + line.baseline}
            fontFamily={cssFontFamily(layout.file)}
            fontSize={fontSize}
            letterSpacing={layout.letterSpacing / PT_PER_MM}
            fill={color}
            style={{ fontKerning: "none", whiteSpace: "pre" }}
          >
            {line.text}
          </text>
          {element.style.underline && line.text && (
            <line
              x1={element.x + line.x}
              x2={element.x + line.x + line.width}
              y1={element.y + line.baseline + fontSize * 0.12}
              y2={element.y + line.baseline + fontSize * 0.12}
              stroke={color}
              strokeWidth={Math.max(fontSize * 0.06, 0.14)}
            />
          )}
        </g>
      ))}
    </g>
  );
}

function LogoShape({ box, color, opacity }: { box: Box; color: string; opacity: number }) {
  return (
    <svg x={box.x} y={box.y} width={box.width} height={box.height} viewBox="80 0 1375 300" opacity={opacity}>
      <g transform="translate(0,430) scale(0.1,-0.1)">
        {[...LOGO_LETTER_PATHS, ...LOGO_FRAME_PATHS, LOGO_CUBE_PATH].map((d) => (
          <path key={d} d={d} fill={color} />
        ))}
        <path d={LOGO_TOP_PATH} fill={LOGO_TOP_COLOR} />
        <path d={LOGO_BOTTOM_PATH} fill={LOGO_BOTTOM_COLOR} />
      </g>
    </svg>
  );
}

function QrShape({ element, context, place }: { element: QrElement; context: CertificateContext | null; place?: number }) {
  const layout = layoutQr(element, context);
  if (layout.state !== "ok") {
    return (
      <g>
        <rect x={element.x} y={element.y} width={element.width} height={element.height} fill="#f2f4f7" stroke={layout.state === "error" ? OVERFLOW : "#b9bdc6"} strokeWidth={0.3} strokeDasharray="1.5 1" />
        <text x={element.x + element.width / 2} y={element.y + element.height / 2 + 1.2} textAnchor="middle" fontSize={3} fill="#5b616e">
          {layout.state === "error" ? "Túl hosszú" : "Nincs adat"}
        </text>
      </g>
    );
  }
  const { box } = layout;
  return (
    <g opacity={element.opacity}>
      {element.background && <rect x={box.x} y={box.y} width={box.width} height={box.height} fill={element.background} />}
      <path
        d={layout.matrix.path}
        transform={`translate(${box.x + element.quietZone * layout.module} ${box.y + element.quietZone * layout.module}) scale(${layout.module})`}
        fill={resolveColor(element.color, place)}
        shapeRendering="crispEdges"
      />
    </g>
  );
}

export function qrProblem(element: CertificateElement, context: CertificateContext | null): string | null {
  if (element.type !== "qr") return null;
  const layout = layoutQr(element, context);
  if (layout.state === "error") return "a tartalma túl hosszú egy QR-kódhoz";
  if (layout.state === "ok" && layout.tooSmall) {
    return `túl kicsi: ${layout.module.toFixed(2).replace(".", ",")} mm-es pontok, legalább 0,4 mm kell a biztos beolvasáshoz`;
  }
  return null;
}

export function textLayouts(
  elements: CertificateElement[],
  context: CertificateContext | null,
  measurer: Measurer | null,
): Map<string, TextLayout> {
  const layouts = new Map<string, TextLayout>();
  if (!measurer) return layouts;
  for (const element of elements) {
    if (element.type !== "text") continue;
    try {
      layouts.set(element.id, layoutText(element, segmentsText(element.content, context), measurer));
    } catch {}
  }
  return layouts;
}

export function ElementShape({
  element,
  layout,
  context,
  place,
  assetUrls,
}: {
  element: CertificateElement;
  layout?: TextLayout;
  context: CertificateContext | null;
  place?: number;
  assetUrls: Record<string, string>;
}) {
  switch (element.type) {
    case "qr":
      return <QrShape element={element} context={context} place={place} />;
    case "text":
      return layout ? <TextShape element={element} layout={layout} place={place} /> : null;
    case "image": {
      const url = element.assetId ? assetUrls[element.assetId] : undefined;
      if (!url) {
        return (
          <g>
            <rect x={element.x} y={element.y} width={element.width} height={element.height} fill="#f2f4f7" stroke="#b9bdc6" strokeWidth={0.3} strokeDasharray="1.5 1" />
            <text x={element.x + element.width / 2} y={element.y + element.height / 2 + 1.2} textAnchor="middle" fontSize={3.5} fill="#5b616e">
              {element.assetId ? "Kép betöltése…" : "Válassz képet"}
            </text>
          </g>
        );
      }
      return (
        <image
          href={url}
          x={element.x}
          y={element.y}
          width={element.width}
          height={element.height}
          preserveAspectRatio={PRESERVE[element.fit]}
          opacity={element.opacity}
        />
      );
    }
    case "rect":
      return (
        <rect
          x={element.x}
          y={element.y}
          width={element.width}
          height={element.height}
          rx={element.radius}
          fill={element.fill ? resolveColor(element.fill, place) : "none"}
          stroke={element.stroke ? resolveColor(element.stroke, place) : "none"}
          strokeWidth={element.stroke ? element.strokeWidth / PT_PER_MM : 0}
          opacity={element.opacity}
        />
      );
    case "ellipse":
      return (
        <ellipse
          cx={element.x + element.width / 2}
          cy={element.y + element.height / 2}
          rx={element.width / 2}
          ry={element.height / 2}
          fill={element.fill ? resolveColor(element.fill, place) : "none"}
          stroke={element.stroke ? resolveColor(element.stroke, place) : "none"}
          strokeWidth={element.stroke ? element.strokeWidth / PT_PER_MM : 0}
          opacity={element.opacity}
        />
      );
    case "line": {
      const [start, end] = lineEnds(element);
      return (
        <line
          x1={start.x}
          y1={start.y}
          x2={end.x}
          y2={end.y}
          stroke={resolveColor(element.stroke, place)}
          strokeWidth={element.strokeWidth / PT_PER_MM}
          opacity={element.opacity}
        />
      );
    }
    case "logo":
      return <LogoShape box={logoBox(element)} color={resolveColor(element.color, place)} opacity={element.opacity} />;
  }
}

