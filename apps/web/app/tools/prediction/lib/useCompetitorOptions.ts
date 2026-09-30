"use client";

import { useEffect, useState } from "react";
import { clientJson } from "@/app/lib/api";
import type { CompetitorOption } from "../types";

export type CompetitorOptionsStatus = "loading" | "ready" | "error";

export type CompetitorOptions = {
  options: CompetitorOption[];
  status: CompetitorOptionsStatus;
};

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

  if (path === null || state.path !== path) {
    return { options: [], status: "loading" };
  }
  return { options: state.options, status: state.status };
}
