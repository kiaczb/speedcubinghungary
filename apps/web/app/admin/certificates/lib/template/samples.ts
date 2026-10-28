import type { CertificateCompetition, Recipient } from "../../types";
import { recipientContext } from "../pdf/content";
import type { TemplateKind } from "./model";
import type { CertificateContext } from "./variables";

export const SAMPLE_COMPETITION: CertificateCompetition = {
  id: "MintaVerseny2026",
  name: "Minta Verseny 2026",
  startDate: "2026-11-14",
  endDate: "2026-11-15",
  city: "Budapest",
  countryIso2: "HU",
  url: "",
  eventIds: [],
};

const PLACEMENT_SAMPLES: Recipient[] = [
  { name: "Kovács Anna", wcaId: "2019KOVA01", eventId: "333", place: 1, result: "7.84", resultType: "average" },
  { name: "Szabó Dávid", wcaId: "2022SZAB03", eventId: "333oh", place: 2, result: "13.35", resultType: "average" },
  { name: "Kacper Paweł Dworak-Wiśniewski", wcaId: "2020DWOR01", eventId: "555", place: 3, result: "58.03", resultType: "average" },
  { name: "Újonc Ödön", eventId: "333mbf", place: 1, result: "16/16 56:38", resultType: "single" },
];

const PARTICIPATION_SAMPLES: Recipient[] = [
  { name: "Kovács Anna", wcaId: "2019KOVA01" },
  { name: "Kacper Paweł Dworak-Wiśniewski", wcaId: "2020DWOR01" },
  { name: "Yiheng Wang (王艺衡)", wcaId: "2019WANY36" },
  { name: "Újonc Ödön" },
];

export type PreviewSet = {
  label: string;
  contexts: CertificateContext[];
};

export function sampleSet(kind: TemplateKind): PreviewSet {
  const recipients = kind === "placement" ? PLACEMENT_SAMPLES : PARTICIPATION_SAMPLES;
  return {
    label: "Mintaadat",
    contexts: recipients.map((r) => recipientContext(SAMPLE_COMPETITION, r)),
  };
}

export function contextName(context: CertificateContext | undefined): string {
  return context?.values["competitor.fullName"] || context?.values["competitor.name"] || "";
}
