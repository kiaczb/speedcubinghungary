"use client";

import { useEffect, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { FaBroadcastTower } from "react-icons/fa";
import { clientApiUrl } from "@/app/lib/api";
import type { LiveLeg } from "../types";

type LiveBannerProps = {
  year: number;
  live?: LiveLeg;
};

const LiveBanner = ({ year, live }: LiveBannerProps) => {
  const router = useRouter();
  const [connected, setConnected] = useState(false);
  const [, startRefresh] = useTransition();

  useEffect(() => {
    let source: EventSource | null = null;

    const connect = () => {
      source = new EventSource(`${clientApiUrl()}/allrounder/stream?year=${year}`, {
        withCredentials: true,
      });

      source.addEventListener("ready", () => setConnected(true));
      source.addEventListener("update", () => startRefresh(() => router.refresh()));
      source.onerror = () => setConnected(false);
    };

    // Mobil böngészők a háttérbe került lapot felfüggesztik/limitálják, ami a
    // meglévő EventSource kapcsolatot megszakíthatja anélkül, hogy az `onerror`
    // lefutna. A komponens ilyenkor is bent marad a DOM-ban (nincs újbóli
    // mountolás), tehát semmi nem indítaná újra a kapcsolatot — a lap
    // visszatérésekor ezért zárjuk és nyitjuk újra kézzel, és egyúttal
    // azonnal frissítünk is, hogy a kiesett idő alatti változás ne maradjon
    // észrevétlen a következő `update` eseményig.
    const handleVisibilityChange = () => {
      if (document.visibilityState !== "visible") return;

      source?.close();
      connect();
      startRefresh(() => router.refresh());
    };

    connect();
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      source?.close();
      setConnected(false);
    };
  }, [year, router]);

  if (!live) return null;

  return (
    <div className="mt-6 flex flex-wrap items-center gap-x-3 gap-y-1 rounded-card border border-line bg-surface-muted px-4 py-3 text-sm">
      <span className="flex items-center gap-2 font-medium text-foreground">
        <FaBroadcastTower
          className={`h-4 w-4 text-brand-red ${connected ? "animate-pulse" : ""}`}
          aria-hidden
        />
        Élő: {live.label}
      </span>
      <span className="text-muted-foreground">
        A pontok a WCA Live {live.roundCount} követett köréből frissülnek, ezért
        ideiglenesek — a hivatalos WCA-eredmény feltöltésekor még változhatnak.
      </span>
    </div>
  );
};

export default LiveBanner;
