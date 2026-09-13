import { clientApiUrl } from "@/app/lib/api";
import type {
  LearnAlgorithm,
  LearnCase,
  LearnEvent,
  LearnEventNode,
} from "@/app/learn/types";

// Böngészőből induló hívások: a JWT sütit `credentials: "include"` viszi.
// A jogosultságot a backend ellenőrzi (JWT + RBAC admin) — az admin felület
// kapuja csak kényelmi funkció, nem védelem.
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${clientApiUrl()}${path}`, {
    credentials: "include",
    ...init,
  });

  if (!res.ok) {
    const message = (await res.text()).trim();
    throw new Error(message || `A kérés sikertelen (${res.status})`);
  }

  // A DELETE 204-et ad, törzs nélkül — ilyenkor nincs mit JSON-ként olvasni.
  if (res.status === 204) return undefined as T;

  return res.json() as Promise<T>;
}

function jsonBody(payload: unknown): RequestInit {
  return {
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  };
}

export type EventPayload = {
  name: string;
  description: string;
  order: number;
};

export type CasePayload = EventPayload & {
  eventId: string;
};

export type AlgorithmPayload = EventPayload & {
  caseId: string;
  notation: string;
  videoUrl: string;
};

/** A teljes hierarchia egy kérésben — az admin felület ebből épül fel. */
export function fetchLearnTree(): Promise<LearnEventNode[]> {
  return request<LearnEventNode[]>("/learn/tree");
}

export function createEvent(payload: EventPayload): Promise<LearnEvent> {
  return request<LearnEvent>("/learn/events", {
    method: "POST",
    ...jsonBody(payload),
  });
}

export function updateEvent(
  id: string,
  payload: EventPayload,
): Promise<LearnEvent> {
  return request<LearnEvent>(`/learn/events/${id}`, {
    method: "PUT",
    ...jsonBody(payload),
  });
}

export function deleteEvent(id: string): Promise<void> {
  return request<void>(`/learn/events/${id}`, { method: "DELETE" });
}

export function createCase(payload: CasePayload): Promise<LearnCase> {
  return request<LearnCase>("/learn/cases", {
    method: "POST",
    ...jsonBody(payload),
  });
}

export function updateCase(
  id: string,
  payload: CasePayload,
): Promise<LearnCase> {
  return request<LearnCase>(`/learn/cases/${id}`, {
    method: "PUT",
    ...jsonBody(payload),
  });
}

export function deleteCase(id: string): Promise<void> {
  return request<void>(`/learn/cases/${id}`, { method: "DELETE" });
}

export function createAlgorithm(
  payload: AlgorithmPayload,
): Promise<LearnAlgorithm> {
  return request<LearnAlgorithm>("/learn/algorithms", {
    method: "POST",
    ...jsonBody(payload),
  });
}

export function updateAlgorithm(
  id: string,
  payload: AlgorithmPayload,
): Promise<LearnAlgorithm> {
  return request<LearnAlgorithm>(`/learn/algorithms/${id}`, {
    method: "PUT",
    ...jsonBody(payload),
  });
}

export function deleteAlgorithm(id: string): Promise<void> {
  return request<void>(`/learn/algorithms/${id}`, { method: "DELETE" });
}
