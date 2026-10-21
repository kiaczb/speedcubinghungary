"use client";

import { useEffect, useId, useRef } from "react";
import type { PointerEvent } from "react";
import { createPortal } from "react-dom";

type ModalSize = "sm" | "md" | "lg";

type Props = {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
  size?: ModalSize;
  role?: "dialog" | "alertdialog";
  closeDisabled?: boolean;
  hideHeader?: boolean;
  initialFocusRef?: React.RefObject<HTMLElement | null>;
};

const SIZE_CLASS: Record<ModalSize, string> = {
  sm: "max-w-sm",
  md: "max-w-lg",
  lg: "max-w-2xl",
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

  useEffect(() => {
    const { overflow, paddingRight } = document.body.style;
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
    document.body.style.overflow = "hidden";
    if (scrollbarWidth > 0) {
      document.body.style.paddingRight = `${scrollbarWidth}px`;
    }

    const openedFrom = document.activeElement as HTMLElement | null;
    (initialFocusRef?.current ?? panelRef.current)?.focus();

    return () => {
      document.body.style.overflow = overflow;
      document.body.style.paddingRight = paddingRight;
      openedFrom?.focus?.();
    };
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

  if (typeof document === "undefined") return null;

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
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
