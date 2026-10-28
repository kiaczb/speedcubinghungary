"use client";

import { useRef, useState } from "react";
import type { DragEvent, PointerEvent as ReactPointerEvent } from "react";
import type { CertificateDocument, CertificateElement } from "../../../lib/template/model";
import type { Box } from "../../../lib/template/geometry";
import type { Measurer } from "../../../lib/template/textLayout";
import type { CertificateContext } from "../../../lib/template/variables";
import { ElementShape, OVERFLOW, qrProblem, textLayouts } from "../../../components/CertificateShapes";
import { DRAG_TYPE, MIN_SIZE } from "../lib/elements";

type Handle = "n" | "s" | "e" | "w" | "ne" | "nw" | "se" | "sw";
type DragKind = "move" | Handle;

type Drag = {
  id: string;
  kind: DragKind;
  startX: number;
  startY: number;
  box: Box;
  moved: boolean;
};

type Guides = { x: number[]; y: number[] };

const SELECTION = "#16181d";
const HANDLE_PX = 9;
const SNAP_PX = 6;

const round = (value: number) => Math.round(value * 10) / 10;

function handlesFor(element: CertificateElement): Handle[] {
  if (element.type === "line") return element.orientation === "horizontal" ? ["w", "e"] : ["n", "s"];
  return ["nw", "n", "ne", "e", "se", "s", "sw", "w"];
}

function handlePoint(box: Box, handle: Handle) {
  const x = handle.includes("w") ? box.x : handle.includes("e") ? box.x + box.width : box.x + box.width / 2;
  const y = handle.includes("n") ? box.y : handle.includes("s") ? box.y + box.height : box.y + box.height / 2;
  return { x, y };
}

const CURSORS: Record<Handle, string> = {
  n: "ns-resize",
  s: "ns-resize",
  e: "ew-resize",
  w: "ew-resize",
  ne: "nesw-resize",
  sw: "nesw-resize",
  nw: "nwse-resize",
  se: "nwse-resize",
};

export type CanvasProps = {
  document: CertificateDocument;
  context: CertificateContext | null;
  place?: number;
  measurer: Measurer | null;
  assetUrls: Record<string, string>;
  scale: number;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  onEditText: (id: string) => void;
  onTransformStart: () => void;
  onTransform: (id: string, box: Box) => void;
  onTransformEnd: () => void;
  onDropElement: (data: string, at: { x: number; y: number }) => void;
};

export default function CertificateCanvas({
  document,
  context,
  place,
  measurer,
  assetUrls,
  scale,
  selectedId,
  onSelect,
  onEditText,
  onTransformStart,
  onTransform,
  onTransformEnd,
  onDropElement,
}: CanvasProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const drag = useRef<Drag | null>(null);
  const [guides, setGuides] = useState<Guides>({ x: [], y: [] });
  const { page } = document;
  const visible = document.elements.filter((element) => !element.hidden);
  const selected = visible.find((element) => element.id === selectedId) ?? null;
  const layouts = textLayouts(visible, context, measurer);
  const overflowing = visible.filter((element) => {
    const layout = layouts.get(element.id);
    return (layout && (layout.overflow || !layout.covered)) || qrProblem(element, context) !== null;
  });

  const toMm = (clientX: number, clientY: number) => {
    const rect = svgRef.current?.getBoundingClientRect();
    if (!rect) return { x: 0, y: 0 };
    return { x: (clientX - rect.left) / scale, y: (clientY - rect.top) / scale };
  };

  const begin = (event: ReactPointerEvent, element: CertificateElement, kind: DragKind) => {
    if (element.locked) return;
    event.stopPropagation();
    if (event.button !== 0) return;
    onSelect(element.id);
    svgRef.current?.setPointerCapture(event.pointerId);
    drag.current = {
      id: element.id,
      kind,
      startX: event.clientX,
      startY: event.clientY,
      box: { x: element.x, y: element.y, width: element.width, height: element.height },
      moved: false,
    };
  };

  const snapMove = (box: Box, id: string, disabled: boolean) => {
    if (disabled) return { box, guides: { x: [], y: [] } };
    const threshold = SNAP_PX / scale;
    const others = visible.filter((e) => e.id !== id);
    const xs = [0, page.margin, page.width / 2, page.width - page.margin, page.width, ...others.flatMap((o) => [o.x, o.x + o.width / 2, o.x + o.width])];
    const ys = [0, page.margin, page.height / 2, page.height - page.margin, page.height, ...others.flatMap((o) => [o.y, o.y + o.height / 2, o.y + o.height])];
    const best = (candidates: number[], anchors: number[]) => {
      let found: { diff: number; at: number } | null = null;
      for (const c of candidates) {
        for (const a of anchors) {
          const diff = c - a;
          if (Math.abs(diff) <= threshold && (!found || Math.abs(diff) < Math.abs(found.diff))) {
            found = { diff, at: c };
          }
        }
      }
      return found;
    };
    const snapX = best(xs, [box.x, box.x + box.width / 2, box.x + box.width]);
    const snapY = best(ys, [box.y, box.y + box.height / 2, box.y + box.height]);
    return {
      box: { ...box, x: box.x + (snapX?.diff ?? 0), y: box.y + (snapY?.diff ?? 0) },
      guides: { x: snapX ? [snapX.at] : [], y: snapY ? [snapY.at] : [] },
    };
  };

  const onPointerMove = (event: ReactPointerEvent) => {
    const current = drag.current;
    if (!current) return;
    const dx = (event.clientX - current.startX) / scale;
    const dy = (event.clientY - current.startY) / scale;
    if (!current.moved) {
      if (Math.abs(dx) * scale < 2 && Math.abs(dy) * scale < 2) return;
      current.moved = true;
      onTransformStart();
    }

    const start = current.box;
    let box: Box;
    if (current.kind === "move") {
      const snapped = snapMove({ ...start, x: start.x + dx, y: start.y + dy }, current.id, event.altKey);
      box = snapped.box;
      setGuides(snapped.guides);
    } else {
      const kind = current.kind;
      let { x, y, width, height } = start;
      if (kind.includes("e")) width = Math.max(MIN_SIZE, start.width + dx);
      if (kind.includes("s")) height = Math.max(MIN_SIZE, start.height + dy);
      if (kind.includes("w")) {
        width = Math.max(MIN_SIZE, start.width - dx);
        x = start.x + start.width - width;
      }
      if (kind.includes("n")) {
        height = Math.max(MIN_SIZE, start.height - dy);
        y = start.y + start.height - height;
      }
      box = { x, y, width, height };
    }
    onTransform(current.id, { x: round(box.x), y: round(box.y), width: round(box.width), height: round(box.height) });
  };

  const onPointerUp = (event: ReactPointerEvent) => {
    const current = drag.current;
    drag.current = null;
    if (svgRef.current?.hasPointerCapture(event.pointerId)) svgRef.current.releasePointerCapture(event.pointerId);
    setGuides({ x: [], y: [] });
    if (current?.moved) onTransformEnd();
  };

  const onDrop = (event: DragEvent<SVGSVGElement>) => {
    const data = event.dataTransfer.getData(DRAG_TYPE);
    if (!data) return;
    event.preventDefault();
    onDropElement(data, toMm(event.clientX, event.clientY));
  };

  const handleSize = HANDLE_PX / scale;

  return (
    <svg
      ref={svgRef}
      width={page.width * scale}
      height={page.height * scale}
      viewBox={`0 0 ${page.width} ${page.height}`}
      className="block touch-none select-none bg-white shadow-panel"
      onPointerDown={() => onSelect(null)}
      onPointerMove={onPointerMove}
      onPointerUp={onPointerUp}
      onPointerCancel={onPointerUp}
      onDragOver={(event) => {
        if (event.dataTransfer.types.includes(DRAG_TYPE)) event.preventDefault();
      }}
      onDrop={onDrop}
    >
      <rect x={0} y={0} width={page.width} height={page.height} fill={page.background || "#ffffff"} />

      {visible.map((element) => (
        <g
          key={element.id}
          onPointerDown={(event) => begin(event, element, "move")}
          onDoubleClick={() => element.type === "text" && onEditText(element.id)}
          style={{ cursor: element.locked ? "default" : "move" }}
        >
          <rect
            x={element.x - (element.type === "line" ? 0.5 : 0)}
            y={element.y - (element.type === "line" ? 0.5 : 0)}
            width={element.width + (element.type === "line" ? 1 : 0)}
            height={element.height + (element.type === "line" ? 1 : 0)}
            fill="transparent"
          />
          <ElementShape element={element} layout={layouts.get(element.id)} context={context} place={place} assetUrls={assetUrls} />
        </g>
      ))}

      <rect
        x={page.margin}
        y={page.margin}
        width={Math.max(page.width - 2 * page.margin, 0)}
        height={Math.max(page.height - 2 * page.margin, 0)}
        fill="none"
        stroke="#8a8f98"
        strokeOpacity={0.6}
        strokeWidth={0.6 / scale}
        strokeDasharray={`${4 / scale} ${3 / scale}`}
        pointerEvents="none"
      />

      {overflowing.map((element) => (
          <rect
            key={`overflow-${element.id}`}
            x={element.x}
            y={element.y}
            width={element.width}
            height={element.height}
            fill={OVERFLOW}
            fillOpacity={0.06}
            stroke={OVERFLOW}
            strokeWidth={1.5 / scale}
            strokeDasharray={`${4 / scale} ${2 / scale}`}
            pointerEvents="none"
          />
        ))}

      {guides.x.map((x) => (
        <line key={`gx-${x}`} x1={x} x2={x} y1={0} y2={page.height} stroke={OVERFLOW} strokeWidth={1 / scale} pointerEvents="none" />
      ))}
      {guides.y.map((y) => (
        <line key={`gy-${y}`} x1={0} x2={page.width} y1={y} y2={y} stroke={OVERFLOW} strokeWidth={1 / scale} pointerEvents="none" />
      ))}

      {selected && (
        <g>
          <rect
            x={selected.x}
            y={selected.y}
            width={selected.width}
            height={selected.height}
            fill="none"
            stroke={SELECTION}
            strokeWidth={1.2 / scale}
            strokeDasharray={selected.locked ? `${3 / scale} ${3 / scale}` : undefined}
            pointerEvents="none"
          />
          {!selected.locked &&
            handlesFor(selected).map((handle) => {
              const point = handlePoint(selected, handle);
              return (
                <rect
                  key={handle}
                  x={point.x - handleSize / 2}
                  y={point.y - handleSize / 2}
                  width={handleSize}
                  height={handleSize}
                  fill="#ffffff"
                  stroke={SELECTION}
                  strokeWidth={1.2 / scale}
                  style={{ cursor: CURSORS[handle] }}
                  onPointerDown={(event) => begin(event, selected, handle)}
                />
              );
            })}
        </g>
      )}
    </svg>
  );
}
