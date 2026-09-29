"use client";

import { useEffect, useState } from "react";
import { clientJson } from "@/app/lib/api";
import type { CompetitorOption } from "../types";

export type CompetitorOptionsStatus = "loading" | "ready" | "error";

export type CompetitorOptions = {
  options: CompetitorOption[];
  status: CompetitorOptionsStatus;
};

// Egy versenyző-listát tölt be a backendről (psych sheet vagy nevezői
// lista), hogy a CompetitorInput-ok osztozhassanak rajta. Egy dobogós
// versenyszám 3 mezője EGY kérést jelentsen, ne hármat — a WCA API kb. 15
// párhuzamos kérésnél már 429-et ad, és egy tucat versenyszámmal ezt a
// mezőnkénti lekérés simán túllépte. `null` path esetén nem kér le semmit
// (pl. nincs versenyző-típusú extra kérdés).
export function useCompetitorOptions(path: string | null): CompetitorOptions {
  const [state, setState] = useState<CompetitorOptions & { path: string | null }>(
    { path: null, options: [], status: "loading" },
  );

  useEffect(() => {
    if (path === null) return;
    let cancelled = false;

    clientJson<CompetitorOption[]>(path)
      .then((options) => {
        if (!cancelled) setState({ path, options, status: "ready" });
      })
      .catch((error) => {
        console.error("A versenyzőlista lekérése sikertelen:", error);
        if (!cancelled) setState({ path, options: [], status: "error" });
      });

    return () => {
      cancelled = true;
    };
  }, [path]);

  // Az állapot a path-hoz van kötve, így egy másik listára váltáskor nem a
  // régi találatokat mutatjuk "ready"-ként (effectben setState nélkül, lásd
  // a react-hooks/set-state-in-effect szabályt).
  if (path === null || state.path !== path) {
    return { options: [], status: "loading" };
  }
  return { options: state.options, status: state.status };
}
