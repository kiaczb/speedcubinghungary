import { formatCompactDateRange, formatLongDate } from "@/app/lib/date";
import { eventName } from "@/app/lib/events";
import type { CertificateCompetition, CertificateMode, Recipient } from "../../types";
import { latinName, resultTypeLabel } from "../recipients";
import type { CertificateContext } from "../template/variables";

const VOWELS = "aáeéiíoóöőuúüű";

export function withArticle(word: string): string {
  const first = word.trim().charAt(0).toLocaleLowerCase("hu-HU");
  const az = VOWELS.includes(first) || first === "1" || first === "5";
  return `${az ? "az" : "a"} ${word}`;
}

export function printOrder(mode: CertificateMode, recipients: Recipient[]): Recipient[] {
  if (mode !== "podium") return recipients;

  const events: string[] = [];
  for (const r of recipients) {
    if (r.eventId && !events.includes(r.eventId)) events.push(r.eventId);
  }
  return events.flatMap((eventId) =>
    recipients
      .filter((r) => r.eventId === eventId)
      .sort((a, b) => (b.place ?? 0) - (a.place ?? 0)),
  );
}

export function competitionValues(competition: CertificateCompetition): CertificateContext["values"] {
  return {
    "competition.name": competition.name,
    "competition.nameWithArticle": withArticle(competition.name),
    "competition.date": formatCompactDateRange(competition.startDate, competition.endDate),
    "competition.city": competition.city,
    "competition.url": competition.url,
    "issue.date": formatLongDate(new Date()),
  };
}

export function recipientContext(
  competition: CertificateCompetition,
  recipient: Recipient,
): CertificateContext {
  const event = recipient.eventId ? eventName(recipient.eventId) : "";
  return {
    place: recipient.place,
    values: {
      ...competitionValues(competition),
      "competitor.name": latinName(recipient.name),
      "competitor.fullName": recipient.name,
      "competitor.wcaId": recipient.wcaId ?? "",
      "result.event": event,
      "result.eventWithArticle": event ? withArticle(event) : "",
      "result.place": recipient.place ? `${recipient.place}.` : "",
      "result.value": recipient.result ?? "",
      "result.type": recipient.result ? resultTypeLabel(recipient.resultType) : "",
    },
  };
}

export function certificateContexts(
  competition: CertificateCompetition,
  mode: CertificateMode,
  recipients: Recipient[],
): CertificateContext[] {
  return printOrder(mode, recipients).map((recipient) => recipientContext(competition, recipient));
}
