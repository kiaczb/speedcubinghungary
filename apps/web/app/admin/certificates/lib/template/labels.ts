import type { TemplateKind } from "./model";
import { PAGE_PRESETS } from "./model";

export const KIND_LABELS: Record<TemplateKind, { title: string; short: string; description: string }> = {
  placement: {
    title: "Dobogós oklevél",
    short: "Dobogós",
    description: "Helyezésenként egy oklevél, a versenyszám, helyezés és eredmény adatmezőivel.",
  },
  participation: {
    title: "Részvételi oklevél",
    short: "Részvételi",
    description: "Versenyzőnként egy oklevél, a Versenyzők és a Regisztráltak módhoz.",
  },
};

export function pageLabel(width: number, height: number): string {
  const preset = PAGE_PRESETS.find(
    (p) => Math.abs(p.width - width) < 0.5 && Math.abs(p.height - height) < 0.5,
  );
  return preset ? preset.label : `${Math.round(width)} × ${Math.round(height)} mm`;
}
