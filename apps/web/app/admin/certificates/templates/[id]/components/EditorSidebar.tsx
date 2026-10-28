"use client";

import type { DragEvent, ReactNode } from "react";
import type { IconType } from "react-icons";
import {
  FaCube,
  FaFont,
  FaLink,
  FaLock,
  FaMapMarkerAlt,
  FaMedal,
  FaMinus,
  FaQrcode,
  FaRegAddressCard,
  FaRegCalendar,
  FaRegCalendarCheck,
  FaRegChartBar,
  FaRegCircle,
  FaRegEye,
  FaRegEyeSlash,
  FaRegFlag,
  FaRegIdCard,
  FaRegImage,
  FaRegSquare,
  FaRegUser,
  FaStar,
  FaStopwatch,
} from "react-icons/fa";
import type { CertificateAsset, CertificateElement, TemplateKind } from "../../../lib/template/model";
import { VARIABLE_GROUPS, variablesFor } from "../../../lib/template/variables";
import type { VariableKey } from "../../../lib/template/variables";
import { DRAG_TYPE, ELEMENT_LABELS } from "../lib/elements";
import type { NewElement } from "../lib/elements";
import AssetLibrary from "./AssetLibrary";

const VARIABLE_ICONS: Record<VariableKey, IconType> = {
  "competitor.name": FaRegUser,
  "competitor.fullName": FaRegAddressCard,
  "competitor.wcaId": FaRegIdCard,
  "competition.name": FaRegFlag,
  "competition.nameWithArticle": FaRegFlag,
  "competition.date": FaRegCalendar,
  "competition.city": FaMapMarkerAlt,
  "competition.url": FaLink,
  "result.event": FaCube,
  "result.eventWithArticle": FaCube,
  "result.place": FaMedal,
  "result.value": FaStopwatch,
  "result.type": FaRegChartBar,
  "issue.date": FaRegCalendarCheck,
};

const ELEMENTS: { spec: NewElement; icon: ReactNode }[] = [
  { spec: { type: "text" }, icon: <FaFont className="h-3.5 w-3.5" /> },
  { spec: { type: "image", assetId: null }, icon: <FaRegImage className="h-3.5 w-3.5" /> },
  { spec: { type: "logo" }, icon: <FaStar className="h-3.5 w-3.5" /> },
  { spec: { type: "rect" }, icon: <FaRegSquare className="h-3.5 w-3.5" /> },
  { spec: { type: "ellipse" }, icon: <FaRegCircle className="h-3.5 w-3.5" /> },
  { spec: { type: "line" }, icon: <FaMinus className="h-3.5 w-3.5" /> },
  { spec: { type: "qr" }, icon: <FaQrcode className="h-3.5 w-3.5" /> },
];

function Heading({ children }: { children: ReactNode }) {
  return <h3 className="text-xs font-semibold text-muted-foreground">{children}</h3>;
}

export default function EditorSidebar({
  kind,
  elements,
  selectedId,
  assets,
  assetUrls,
  onAdd,
  onSelect,
  onToggle,
  onUpload,
  onDeleteAsset,
}: {
  kind: TemplateKind;
  elements: CertificateElement[];
  selectedId: string | null;
  assets: CertificateAsset[];
  assetUrls: Record<string, string>;
  onAdd: (spec: NewElement) => void;
  onSelect: (id: string) => void;
  onToggle: (id: string, flag: "hidden" | "locked") => void;
  onUpload: (file: File) => Promise<void>;
  onDeleteAsset: (asset: CertificateAsset) => void;
}) {
  const variables = variablesFor(kind);
  const drag = (spec: NewElement) => (event: DragEvent) => {
    event.dataTransfer.setData(DRAG_TYPE, JSON.stringify(spec));
    event.dataTransfer.effectAllowed = "copy";
  };

  return (
    <div className="flex flex-col divide-y divide-line">
      <section className="flex flex-col gap-3 p-4">
        <Heading>Elemek</Heading>
        <div className="grid grid-cols-2 gap-1.5">
          {ELEMENTS.map(({ spec, icon }) => (
            <button
              key={spec.type}
              type="button"
              draggable
              onDragStart={drag(spec)}
              onClick={() => onAdd(spec)}
              className="flex items-center gap-2 rounded-control border border-line bg-surface px-2.5 py-2 text-left text-sm text-foreground transition-colors duration-200 ease-smooth hover:border-foreground"
            >
              <span className="text-muted-foreground">{icon}</span>
              {ELEMENT_LABELS[spec.type]}
            </button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">Kattints vagy húzd a lapra.</p>
      </section>

      <section className="flex flex-col gap-2 p-4">
        <Heading>Adatmezők</Heading>
        {VARIABLE_GROUPS.map((group) => {
          const inGroup = variables.filter((v) => v.group === group);
          if (inGroup.length === 0) return null;
          return (
            <div key={group}>
              <p className="pb-1 text-xs text-subtle-foreground">{group}</p>
              <ul className="flex flex-col">
                {inGroup.map((v) => {
                  const spec: NewElement = { type: "text", variable: v.key };
                  const Icon = VARIABLE_ICONS[v.key];
                  return (
                    <li key={v.key}>
                      <button
                        type="button"
                        draggable
                        onDragStart={drag(spec)}
                        onClick={() => onAdd(spec)}
                        className="flex w-full cursor-grab items-center gap-2.5 rounded-md px-2 py-1 text-left text-sm text-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted active:cursor-grabbing"
                      >
                        <Icon aria-hidden className="h-3.5 w-3.5 shrink-0 text-subtle-foreground" />
                        {v.label}
                      </button>
                    </li>
                  );
                })}
              </ul>
            </div>
          );
        })}
        <p className="text-xs text-muted-foreground">
          Egy szövegen belül is keverheted a sima szöveget és az adatmezőket: jelölj ki egy szöveget, és a jobb oldalon szúrd be.
        </p>
      </section>

      <section className="flex flex-col gap-3 p-4">
        <Heading>Képtár</Heading>
        <AssetLibrary
          assets={assets}
          assetUrls={assetUrls}
          draggable
          onPick={(asset) => onAdd({ type: "image", assetId: asset.id, aspect: asset.width / asset.height })}
          onUpload={onUpload}
          onDelete={onDeleteAsset}
        />
      </section>

      <section className="flex flex-col gap-2 p-4">
        <Heading>Rétegek</Heading>
        {elements.length === 0 ? (
          <p className="text-xs text-muted-foreground">Még üres a lap.</p>
        ) : (
          <ul className="flex flex-col gap-0.5">
            {[...elements].reverse().map((element) => (
              <li key={element.id}>
                <div
                  className={`flex items-center gap-1 rounded-md px-2 py-1 text-sm ${
                    element.id === selectedId ? "bg-foreground text-background" : "text-foreground hover:bg-surface-muted"
                  }`}
                >
                  <button type="button" onClick={() => onSelect(element.id)} className={`min-w-0 flex-1 truncate text-left ${element.hidden ? "opacity-50" : ""}`}>
                    {element.name}
                  </button>
                  <button
                    type="button"
                    aria-label={element.hidden ? `${element.name} megjelenítése` : `${element.name} elrejtése`}
                    onClick={() => onToggle(element.id, "hidden")}
                    className="flex h-6 w-6 items-center justify-center rounded opacity-70 hover:opacity-100"
                  >
                    {element.hidden ? <FaRegEyeSlash className="h-3 w-3" /> : <FaRegEye className="h-3 w-3" />}
                  </button>
                  <button
                    type="button"
                    aria-label={element.locked ? `${element.name} zárolásának feloldása` : `${element.name} zárolása`}
                    onClick={() => onToggle(element.id, "locked")}
                    className={`flex h-6 w-6 items-center justify-center rounded hover:opacity-100 ${element.locked ? "opacity-100" : "opacity-30"}`}
                  >
                    <FaLock className="h-3 w-3" />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
