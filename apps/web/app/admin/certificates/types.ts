export type CertificateMode = "podium" | "competitors" | "registrants";

export type RecipientSource = "registrations" | "results" | "live" | "none";

export type EventStatus = "finished" | "running" | "pending";

export type CertificateCompetition = {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  city: string;
  countryIso2: string;
  url: string;
  eventIds: string[];
  resultsPostedAt?: string;
};

export type CertificateEvent = {
  eventId: string;
  status: EventStatus;
};

export type Recipient = {
  name: string;
  wcaId?: string;
  countryIso2?: string;
  eventId?: string;
  place?: number;
  result?: string;
  resultType?: "single" | "average";
};

export type RecipientList = {
  competition: CertificateCompetition;
  mode: CertificateMode;
  source: RecipientSource;
  fetchedAt?: string;
  events: CertificateEvent[];
  recipients: Recipient[];
};
