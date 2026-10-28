import type { CertificateMode, Recipient } from "../types";
import type { TemplateKind } from "./template/model";

export const MODES: { mode: CertificateMode; label: string; description: string }[] = [
  {
    mode: "podium",
    label: "Dobogósok",
    description:
      "Versenyszámonként a döntő 1–3. helyezettje, érvényes eredménnyel. Holtversenynél mindketten ugyanazt a helyezést kapják.",
  },
  {
    mode: "competitors",
    label: "Versenyzők",
    description: "Mindenki, akinek legalább egy eredménye van a versenyen.",
  },
  {
    mode: "registrants",
    label: "Regisztráltak",
    description:
      "Az elfogadott nevezők a WCA-ról. A meg nem jelent versenyzők is benne vannak.",
  },
];

export function modesFor(kind: TemplateKind) {
  return MODES.filter((m) => templateKind(m.mode) === kind);
}

export function parseMode(value: string | null, kind: TemplateKind): CertificateMode {
  const modes = modesFor(kind);
  return modes.find((m) => m.mode === value)?.mode ?? modes[0].mode;
}

const COLLATOR = new Intl.Collator("hu", { sensitivity: "base" });

export function sortByName(recipients: Recipient[]): Recipient[] {
  return [...recipients].sort((a, b) => COLLATOR.compare(a.name, b.name));
}

export function recipientKey(recipient: Recipient): string {
  return [recipient.eventId ?? "", recipient.place ?? "", recipient.wcaId || recipient.name].join("|");
}

export function latinName(name: string): string {
  return name.replace(/\s*\([^)]*\)\s*$/, "").trim() || name;
}

export function placeLabel(place: number): string {
  return `${place}.`;
}

export function resultTypeLabel(type?: "single" | "average"): string {
  return type === "average" ? "átlag" : "legjobb";
}

export function templateKind(mode: CertificateMode): TemplateKind {
  return mode === "podium" ? "placement" : "participation";
}
