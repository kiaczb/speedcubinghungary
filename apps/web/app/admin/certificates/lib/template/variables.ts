import type { TemplateKind, TextSegment } from "./model";

export type VariableKey =
  | "competitor.name"
  | "competitor.fullName"
  | "competitor.wcaId"
  | "competition.name"
  | "competition.nameWithArticle"
  | "competition.date"
  | "competition.city"
  | "competition.url"
  | "result.event"
  | "result.eventWithArticle"
  | "result.place"
  | "result.value"
  | "result.type"
  | "issue.date";

export type VariableGroup = "Versenyző" | "Verseny" | "Helyezés" | "Egyéb";

export type VariableDefinition = {
  key: VariableKey;
  label: string;
  group: VariableGroup;
  placementOnly?: boolean;
};

export const VARIABLES: VariableDefinition[] = [
  { key: "competitor.name", label: "Versenyző neve", group: "Versenyző" },
  { key: "competitor.fullName", label: "Teljes WCA-név", group: "Versenyző" },
  { key: "competitor.wcaId", label: "WCA ID", group: "Versenyző" },
  { key: "competition.name", label: "Verseny neve", group: "Verseny" },
  { key: "competition.nameWithArticle", label: "a/az + Verseny neve", group: "Verseny" },
  { key: "competition.date", label: "Verseny dátuma", group: "Verseny" },
  { key: "competition.city", label: "Város", group: "Verseny" },
  { key: "competition.url", label: "Verseny WCA-oldala", group: "Verseny" },
  { key: "result.event", label: "Versenyszám", group: "Helyezés", placementOnly: true },
  { key: "result.eventWithArticle", label: "a/az + Versenyszám", group: "Helyezés", placementOnly: true },
  { key: "result.place", label: "Helyezés (1.)", group: "Helyezés", placementOnly: true },
  { key: "result.value", label: "Eredmény", group: "Helyezés", placementOnly: true },
  { key: "result.type", label: "Eredmény típusa", group: "Helyezés", placementOnly: true },
  { key: "issue.date", label: "Kiállítás dátuma", group: "Egyéb" },
];

export const VARIABLE_GROUPS: VariableGroup[] = ["Versenyző", "Verseny", "Helyezés", "Egyéb"];

export type CertificateContext = {
  values: Partial<Record<VariableKey, string>>;
  place?: number;
};

export function variableLabel(key: VariableKey): string {
  return VARIABLES.find((v) => v.key === key)?.label ?? key;
}

export function variablesFor(kind: TemplateKind): VariableDefinition[] {
  return VARIABLES.filter((v) => kind === "placement" || !v.placementOnly);
}

export function isKnownVariable(key: string, kind: TemplateKind): boolean {
  return variablesFor(kind).some((v) => v.key === key);
}

export function hasMissingValue(segments: TextSegment[], context: CertificateContext): boolean {
  return segments.some((segment) => "variable" in segment && !context.values[segment.variable]);
}

export function segmentsText(
  segments: TextSegment[],
  context: CertificateContext | null,
): string {
  return segments
    .map((segment) =>
      "variable" in segment
        ? context
          ? (context.values[segment.variable] ?? "")
          : `[${variableLabel(segment.variable)}]`
        : segment.text,
    )
    .join("");
}
