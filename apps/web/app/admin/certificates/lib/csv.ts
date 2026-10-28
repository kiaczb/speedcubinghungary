import { eventName } from "@/app/lib/events";
import { formatDateRange } from "@/app/lib/date";
import { countryName } from "../../stats/lib/format";
import type { CertificateCompetition, CertificateMode, Recipient } from "../types";
import { downloadBlob } from "./download";
import { latinName, placeLabel, resultTypeLabel } from "./recipients";

const SEPARATOR = ";";

function cell(value: string): string {
  return /[;"\r\n]/.test(value) ? `"${value.replaceAll('"', '""')}"` : value;
}

export function recipientsCsv(
  competition: CertificateCompetition,
  mode: CertificateMode,
  recipients: Recipient[],
): string {
  const podium = mode === "podium";
  const header = [
    "Név",
    "WCA-név",
    "WCA ID",
    "Ország",
    ...(podium ? ["Versenyszám", "Helyezés", "Eredmény", "Eredmény típusa"] : []),
    "Verseny",
    "Dátum",
    "Helyszín",
  ];
  const date = formatDateRange(competition.startDate, competition.endDate);

  const rows = recipients.map((r) => [
    latinName(r.name),
    r.name,
    r.wcaId ?? "",
    countryName(r.countryIso2),
    ...(podium
      ? [
          r.eventId ? eventName(r.eventId) : "",
          r.place ? placeLabel(r.place) : "",
          r.result ?? "",
          resultTypeLabel(r.resultType),
        ]
      : []),
    competition.name,
    date,
    competition.city,
  ]);

  return [header, ...rows].map((row) => row.map(cell).join(SEPARATOR)).join("\r\n");
}

export function downloadCsv(filename: string, content: string) {
  downloadBlob(filename, new Blob(["﻿", content], { type: "text/csv;charset=utf-8" }));
}
