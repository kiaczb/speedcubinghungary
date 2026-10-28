"use client";

import { useState } from "react";
import type { ReactNode } from "react";
import { controlClass } from "@/app/components/ui/Field";
import { MEDAL } from "../../../lib/template/model";

const compactControl = `${controlClass} px-2.5 py-1.5`;

export function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-3 border-t border-line px-4 py-4 first:border-t-0">
      <h3 className="text-xs font-semibold text-muted-foreground">{title}</h3>
      {children}
    </section>
  );
}

export function Row({ label, htmlFor, children }: { label: string; htmlFor?: string; children: ReactNode }) {
  return (
    <div className="grid grid-cols-[6.5rem_1fr] items-center gap-2">
      <label htmlFor={htmlFor} className="text-sm text-muted-foreground">
        {label}
      </label>
      <div className="min-w-0">{children}</div>
    </div>
  );
}

function format(value: number, decimals: number): string {
  return String(Math.round(value * 10 ** decimals) / 10 ** decimals).replace(".", ",");
}

export function NumberInput({
  id,
  value,
  onChange,
  min,
  max,
  step = 1,
  decimals = 1,
  unit,
  ariaLabel,
}: {
  id?: string;
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  step?: number;
  decimals?: number;
  unit?: string;
  ariaLabel?: string;
}) {
  const [draft, setDraft] = useState<string | null>(null);

  const commit = (raw: string) => {
    const parsed = Number.parseFloat(raw.replace(",", "."));
    if (!Number.isFinite(parsed)) return;
    const clamped = Math.min(max ?? Infinity, Math.max(min ?? -Infinity, parsed));
    onChange(Math.round(clamped * 10 ** decimals) / 10 ** decimals);
  };

  return (
    <div className="relative">
      <input
        id={id}
        aria-label={ariaLabel}
        inputMode="decimal"
        value={draft ?? format(value, decimals)}
        onFocus={(event) => {
          setDraft(format(value, decimals));
          event.target.select();
        }}
        onChange={(event) => {
          setDraft(event.target.value);
          commit(event.target.value);
        }}
        onBlur={() => setDraft(null)}
        onKeyDown={(event) => {
          if (event.key !== "ArrowUp" && event.key !== "ArrowDown") return;
          event.preventDefault();
          const next = value + (event.key === "ArrowUp" ? step : -step) * (event.shiftKey ? 10 : 1);
          const clamped = Math.min(max ?? Infinity, Math.max(min ?? -Infinity, next));
          setDraft(format(clamped, decimals));
          onChange(Math.round(clamped * 10 ** decimals) / 10 ** decimals);
        }}
        className={`${compactControl} tabular pr-9 font-mono`}
      />
      {unit && (
        <span className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-xs text-subtle-foreground">
          {unit}
        </span>
      )}
    </div>
  );
}

export function TextInput({
  id,
  value,
  onChange,
  maxLength,
}: {
  id?: string;
  value: string;
  onChange: (value: string) => void;
  maxLength?: number;
}) {
  return (
    <input id={id} value={value} maxLength={maxLength} onChange={(e) => onChange(e.target.value)} className={compactControl} />
  );
}

export function Select<T extends string>({
  id,
  value,
  options,
  onChange,
}: {
  id?: string;
  value: T;
  options: { value: T; label: string }[];
  onChange: (value: T) => void;
}) {
  return (
    <select id={id} value={value} onChange={(e) => onChange(e.target.value as T)} className={compactControl}>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

export function ToggleGroup<T extends string>({
  value,
  options,
  onChange,
  label,
}: {
  value: T | T[];
  options: { value: T; label: ReactNode; title: string; disabled?: boolean }[];
  onChange: (value: T) => void;
  label: string;
}) {
  const active = Array.isArray(value) ? value : [value];
  return (
    <div role="group" aria-label={label} className="inline-flex rounded-control border border-line bg-surface p-0.5">
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          title={option.title}
          aria-label={option.title}
          aria-pressed={active.includes(option.value)}
          disabled={option.disabled}
          onClick={() => onChange(option.value)}
          className={`flex h-7 min-w-8 items-center justify-center rounded-md px-2 text-sm transition-colors duration-200 ease-smooth disabled:opacity-40 ${
            active.includes(option.value)
              ? "bg-foreground text-background"
              : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
          }`}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}

export function Checkbox({ checked, onChange, children }: { checked: boolean; onChange: (checked: boolean) => void; children: ReactNode }) {
  return (
    <label className="flex cursor-pointer items-center gap-2 text-sm text-foreground">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} className="h-4 w-4 accent-foreground" />
      {children}
    </label>
  );
}

const HEX = /^#[0-9a-fA-F]{6}$/;

export function ColorInput({
  id,
  value,
  onChange,
  allowMedal,
  allowNone,
}: {
  id?: string;
  value: string | null;
  onChange: (value: string | null) => void;
  allowMedal?: boolean;
  allowNone?: boolean;
}) {
  const [draft, setDraft] = useState<string | null>(null);
  const medal = value === MEDAL;
  const none = value === null;
  const hex = value && HEX.test(value) ? value : "#16181d";

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <input
          type="color"
          aria-label="Szín választása"
          value={hex}
          disabled={medal || none}
          onChange={(e) => onChange(e.target.value)}
          className="h-8 w-10 shrink-0 cursor-pointer rounded-md border border-line bg-surface p-0.5 disabled:cursor-not-allowed disabled:opacity-40"
        />
        <input
          id={id}
          value={draft ?? (medal ? "helyezés szerint" : none ? "nincs" : hex)}
          disabled={medal || none}
          onFocus={() => setDraft(hex)}
          onBlur={() => setDraft(null)}
          onChange={(e) => {
            setDraft(e.target.value);
            if (HEX.test(e.target.value)) onChange(e.target.value.toLowerCase());
          }}
          className={`${compactControl} tabular font-mono disabled:opacity-60`}
        />
      </div>
      {(allowMedal || allowNone) && (
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          {allowNone && (
            <Checkbox checked={none} onChange={(checked) => onChange(checked ? null : "#16181d")}>
              Nincs
            </Checkbox>
          )}
          {allowMedal && (
            <Checkbox checked={medal} onChange={(checked) => onChange(checked ? MEDAL : "#16181d")}>
              Helyezés szerint (arany/ezüst/bronz)
            </Checkbox>
          )}
        </div>
      )}
    </div>
  );
}
