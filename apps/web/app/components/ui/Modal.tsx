"use client";

import { useEffect, useId, useRef } from "react";
import type { PointerEvent } from "react";
import { createPortal } from "react-dom";
import { useBodyScrollLock } from "@/app/lib/useBodyScrollLock";

/**
 * A modális párbeszédek közös váza: háttér, panel, fejléc, Escape,
 * háttérkattintás, görgetészár — és a FÓKUSZCSAPDA.
 *
 * ⚠️ A fókuszcsapda az, ami eddig egyik másolatban sem volt benne. Enélkül a
 * Tab kiviszi a fókuszt a párbeszédből a mögötte lévő — vizuálisan letakart,
 * de továbbra is elérhető — oldalra: a billentyűzettel vagy képernyőolvasóval
 * dolgozó ember "elveszti" a párbeszédet, és olyan gombokat üt le, amiket nem
 * is lát. Ugyanitt tartjuk a nyitáskori fókuszálást és a bezáráskori
 * visszaadást is, mert a három egy dolog: hol van a fókusz, amíg a párbeszéd
 * nyitva van.
 */

type ModalSize = "sm" | "md" | "lg";

type Props = {
  /** A párbeszéd címe. Fejléc nélkül is ez lesz a hozzáférhető neve. */
  title: string;
  onClose: () => void;
  children: React.ReactNode;
  size?: ModalSize;
  /** A megerősítő párbeszédek `alertdialog`-ot használnak. */
  role?: "dialog" | "alertdialog";
  /**
   * Mentés közben a bezárás minden útja tiltott (Escape, háttérkattintás, ×) —
   * különben a felhasználó egy futó kérés közben csukná be az űrlapot.
   */
  closeDisabled?: boolean;
  /** A megerősítő párbeszédnek nincs fejlécsávja: a címe a törzsben áll. */
  hideHeader?: boolean;
  /** Nyitáskor ide kerül a fókusz; enélkül magára a panelre. */
  initialFocusRef?: React.RefObject<HTMLElement | null>;
};

const SIZE_CLASS: Record<ModalSize, string> = {
  sm: "max-w-sm", // megerősítés
  md: "max-w-lg", // olvasnivaló (pl. szabályok)
  lg: "max-w-2xl", // űrlap
};

const FOCUSABLE_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  '[tabindex]:not([tabindex="-1"])',
].join(", ");

export default function Modal({
  title,
  onClose,
  children,
  size = "md",
  role = "dialog",
  closeDisabled = false,
  hideHeader = false,
  initialFocusRef,
}: Props) {
  const titleId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const pressedOnBackdrop = useRef(false);

  useBodyScrollLock(true);

  // Csak a beilleszkedéskor/leszedéskor fut: a fókuszt egyszer állítjuk be, és
  // a bezárás után oda adjuk vissza, ahonnan a párbeszéd indult (tipikusan a
  // megnyitó gomb). A closeDisabled váltakozása nem indíthatja újra, különben
  // a mentés elindítása elugrasztaná a fókuszt.
  useEffect(() => {
    const openedFrom = document.activeElement as HTMLElement | null;

    (initialFocusRef?.current ?? panelRef.current)?.focus();

    return () => openedFrom?.focus?.();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (!closeDisabled) onClose();
        return;
      }

      if (event.key !== "Tab") return;

      const panel = panelRef.current;
      if (!panel) return;

      const focusable = Array.from(
        panel.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
      );

      // Nincs mit fókuszálni (pl. minden mező disabled mentés közben): a
      // fókusz akkor is a panelen marad, nem szökik ki az oldalra.
      if (focusable.length === 0) {
        event.preventDefault();
        panel.focus();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const active = document.activeElement;

      if (event.shiftKey && (active === first || active === panel)) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && active === last) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onClose, closeDisabled]);

  // A párbeszéd csak felhasználói művelet után renderelődik, de a portál
  // szerveroldalon nem működne — ez a kapu teszi biztonságossá.
  if (typeof document === "undefined") return null;

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm dark:bg-black/70"
      // ⚠️ Csak akkor zárunk, ha a gomb a háttéren ment le ÉS ott is jött fel.
      // Egy szövegkijelölés a panelben kezdődhet és a háttéren érhet véget; a
      // böngésző ilyenkor a közös ősnek, a háttérnek küldi a clicket — és egy
      // félig kitöltött űrlap szó nélkül eltűnne.
      onPointerDown={(event: PointerEvent<HTMLDivElement>) => {
        pressedOnBackdrop.current = event.target === event.currentTarget;
      }}
      onClick={(event) => {
        const onBackdrop =
          pressedOnBackdrop.current && event.target === event.currentTarget;
        pressedOnBackdrop.current = false;
        if (onBackdrop && !closeDisabled) onClose();
      }}
      role="presentation"
    >
      <div
        ref={panelRef}
        tabIndex={-1}
        className={`max-h-[90vh] w-full ${SIZE_CLASS[size]} overflow-y-auto rounded-panel border border-line bg-surface shadow-panel`}
        role={role}
        aria-modal="true"
        {...(hideHeader ? { "aria-label": title } : { "aria-labelledby": titleId })}
      >
        {!hideHeader && (
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-line bg-surface/95 px-6 py-4 backdrop-blur-sm">
            <h2 id={titleId} className="text-lg font-bold">
              {title}
            </h2>
            <button
              type="button"
              onClick={onClose}
              disabled={closeDisabled}
              aria-label="Bezárás"
              className="-mr-2 -mt-1 inline-flex h-8 w-8 items-center justify-center rounded-control text-xl leading-none text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
            >
              ×
            </button>
          </div>
        )}

        {children}
      </div>
    </div>,
    document.body,
  );
}
