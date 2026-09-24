"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { FaSyncAlt } from "react-icons/fa";
import { useUser } from "@/app/components/UserProvider";
import Button from "@/app/components/ui/Button";
import { clientFetch } from "@/app/lib/api";

type RefreshButtonProps = {
  year: number;
};

/**
 * Az évad újratöltése a WCA API-ról, a gyorsítótár lejáratát megkerülve.
 * Akkor kell, ha egy állomás eredménye a WCA-nál utólag módosult, vagy épp
 * most töltötték fel — magától a lekérdezés csak órák múlva ismétlődne.
 *
 * Csak adminnak jelenik meg, mert forgalmat generál a WCA felé; a jogosultságot
 * a backend külön is ellenőrzi (a kliensoldali rejtés csak kényelmi).
 */
const RefreshButton = ({ year }: RefreshButtonProps) => {
  const { user } = useUser();
  const router = useRouter();

  const [refreshing, setRefreshing] = useState(false);
  const [failed, setFailed] = useState(false);
  const [reloading, startReload] = useTransition();

  if (!user?.roles.includes("admin")) return null;

  const busy = refreshing || reloading;

  const handleClick = async () => {
    setRefreshing(true);
    setFailed(false);

    try {
      await clientFetch(`/allrounder/refresh?year=${year}`, { method: "POST" });

      startReload(() => router.refresh());
    } catch (error) {
      console.error("Az All-Rounder évad frissítése sikertelen:", error);
      setFailed(true);
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={handleClick}
      disabled={busy}
      icon={<FaSyncAlt className={`h-3.5 w-3.5 ${busy ? "animate-spin" : ""}`} />}
      title="Eredmények újratöltése a WCA API-ról"
    >
      {failed ? "Nem sikerült" : busy ? "Frissítés…" : "Frissítés"}
    </Button>
  );
};

export default RefreshButton;
