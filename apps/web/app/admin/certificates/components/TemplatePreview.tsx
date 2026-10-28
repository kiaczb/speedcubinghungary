"use client";

import { useEffect, useState } from "react";
import type { CertificateDocument } from "../lib/template/model";
import { documentFontFiles } from "../lib/template/fonts";
import type { Measurer } from "../lib/template/textLayout";
import type { CertificateContext } from "../lib/template/variables";
import { ElementShape, textLayouts } from "./CertificateShapes";

export default function TemplatePreview({
  document,
  context,
  label,
  className,
}: {
  document: CertificateDocument;
  context: CertificateContext | null;
  label: string;
  className?: string;
}) {
  const [measurer, setMeasurer] = useState<Measurer | null>(null);
  const [assetUrls, setAssetUrls] = useState<Record<string, string>>({});
  const { page } = document;
  const visible = document.elements.filter((element) => !element.hidden);

  const fontKey = documentFontFiles(document).sort().join("|");
  useEffect(() => {
    let cancelled = false;
    import("../lib/template/resources")
      .then(({ editorMeasurer }) => editorMeasurer(fontKey ? fontKey.split("|") : []))
      .then((loaded) => {
        if (!cancelled) setMeasurer(loaded);
      })
      .catch((error) => console.error("A betűkészlet nem tölthető be:", error));
    return () => {
      cancelled = true;
    };
  }, [fontKey]);

  const assetKey = [...new Set(visible.flatMap((e) => (e.type === "image" && e.assetId ? [e.assetId] : [])))]
    .sort()
    .join("|");
  useEffect(() => {
    let cancelled = false;
    import("../lib/template/resources").then(({ assetObjectUrl }) => {
      for (const id of assetKey ? assetKey.split("|") : []) {
        assetObjectUrl(id)
          .then((url) => {
            if (!cancelled) setAssetUrls((current) => (current[id] === url ? current : { ...current, [id]: url }));
          })
          .catch((error) => console.error("A kép nem tölthető be:", id, error));
      }
    });
    return () => {
      cancelled = true;
    };
  }, [assetKey]);

  const layouts = textLayouts(visible, context, measurer);

  return (
    <svg
      viewBox={`0 0 ${page.width} ${page.height}`}
      role="img"
      aria-label={label}
      className={`block bg-white ${className ?? ""}`}
      style={{ aspectRatio: `${page.width} / ${page.height}` }}
    >
      <rect x={0} y={0} width={page.width} height={page.height} fill={page.background || "#ffffff"} />
      {visible.map((element) => (
        <ElementShape
          key={element.id}
          element={element}
          layout={layouts.get(element.id)}
          context={context}
          place={context?.place}
          assetUrls={assetUrls}
        />
      ))}
    </svg>
  );
}
