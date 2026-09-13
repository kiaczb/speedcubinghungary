import { serverApiUrl } from "@/app/lib/api";
import type {
  LearnAlgorithmDetail,
  LearnCaseDetail,
  LearnEvent,
  LearnEventDetail,
} from "../types";

// Szerveroldali lekérések a tanulófelület oldalaihoz. A tartalmat admin
// szerkeszti, ezért `no-store`: egy mentés után a következő betöltés már az új
// állapotot mutassa.
async function getJSON<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${serverApiUrl()}${path}`, { cache: "no-store" });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch (error) {
    console.error("Nem sikerült elérni a backendet:", error);
    return null;
  }
}

/** Hálózati hiba és üres lista nem különbözik a listaoldalon — mindkettő üres állapot. */
export async function getLearnEvents(): Promise<LearnEvent[]> {
  return (await getJSON<LearnEvent[]>("/learn/events")) ?? [];
}

export async function getLearnEvent(
  eventId: string,
): Promise<LearnEventDetail | null> {
  return getJSON<LearnEventDetail>(
    `/learn/events/${encodeURIComponent(eventId)}`,
  );
}

export async function getLearnCase(
  caseId: string,
): Promise<LearnCaseDetail | null> {
  return getJSON<LearnCaseDetail>(`/learn/cases/${encodeURIComponent(caseId)}`);
}

export async function getLearnAlgorithm(
  algorithmId: string,
): Promise<LearnAlgorithmDetail | null> {
  return getJSON<LearnAlgorithmDetail>(
    `/learn/algorithms/${encodeURIComponent(algorithmId)}`,
  );
}
