"use client";

import { useState } from "react";
import type { ReactNode } from "react";
import Card from "@/app/components/Cards/Card";
import { RECORD_BADGE_COLORS } from "@/app/posts/lib/recordBadges";
import type { Person } from "../types";
import { countryName, formatNumber } from "../lib/format";

export function StatSection({
  title,
  description,
  children,
}: {
  title: string;
  description?: ReactNode;
  children: ReactNode;
}) {
  return (
    <section className="flex flex-col gap-5">
      <div className="border-b border-line pb-3">
        <h2 className="text-xl font-bold text-foreground">{title}</h2>
        {description && (
          <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
            {description}
          </p>
        )}
      </div>
      {children}
    </section>
  );
}

export function StatBlock({
  title,
  note,
  children,
  className = "",
}: {
  title: string;
  note?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <Card padding="md" className={className}>
      <h3 className="font-semibold text-foreground">{title}</h3>
      {note && (
        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
          {note}
        </p>
      )}
      <div className="mt-4">{children}</div>
    </Card>
  );
}

export function PersonName({ person }: { person: Person }) {
  return (
    <span className="min-w-0">
      <span className="text-foreground">{person.name}</span>
      {person.countryIso2 && person.countryIso2 !== "HU" && (
        <span className="ml-1.5 text-xs text-subtle-foreground">
          {countryName(person.countryIso2)}
        </span>
      )}
    </span>
  );
}

export type Row = {
  key: string;
  label: ReactNode;
  sub?: ReactNode;
  value?: ReactNode;
};

export function Rows({
  rows,
  empty,
  limit = 8,
}: {
  rows: Row[];
  empty: string;
  limit?: number;
}) {
  const [expanded, setExpanded] = useState(false);

  if (rows.length === 0) {
    return <p className="text-sm text-muted-foreground">{empty}</p>;
  }

  const shown = expanded ? rows : rows.slice(0, limit);

  return (
    <div>
      <ul className="flex flex-col divide-y divide-line">
        {shown.map((row) => (
          <li
            key={row.key}
            className="flex items-baseline justify-between gap-4 py-2 text-sm first:pt-0 last:pb-0"
          >
            <span className="min-w-0">
              {row.label}
              {row.sub && (
                <span className="mt-0.5 block text-xs text-muted-foreground">
                  {row.sub}
                </span>
              )}
            </span>
            {row.value !== undefined && (
              <span className="tabular shrink-0 text-right font-mono text-sm text-foreground">
                {row.value}
              </span>
            )}
          </li>
        ))}
      </ul>
      {rows.length > limit && (
        <button
          type="button"
          onClick={() => setExpanded((value) => !value)}
          className="mt-3 rounded-control text-sm font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
        >
          {expanded ? "Kevesebb" : `Mind (${rows.length})`}
        </button>
      )}
    </div>
  );
}

export function KeyFigures({
  figures,
}: {
  figures: { label: string; value: ReactNode; sub?: ReactNode }[];
}) {
  return (
    <dl className="grid grid-cols-2 gap-x-6 gap-y-5 sm:grid-cols-3 lg:grid-cols-6">
      {figures.map((figure) => (
        <div key={figure.label}>
          <dt className="text-xs text-muted-foreground">{figure.label}</dt>
          <dd className="tabular mt-1 font-mono text-2xl font-bold text-foreground">
            {figure.value}
          </dd>
          {figure.sub && (
            <dd className="mt-0.5 text-xs text-subtle-foreground">
              {figure.sub}
            </dd>
          )}
        </div>
      ))}
    </dl>
  );
}

export function BarList({
  items,
  limit = 12,
}: {
  items: { key: string; label: ReactNode; count: number }[];
  limit?: number;
}) {
  const [expanded, setExpanded] = useState(false);
  const max = Math.max(1, ...items.map((item) => item.count));
  const shown = expanded ? items : items.slice(0, limit);

  return (
    <div>
      <ul className="flex flex-col gap-2.5">
        {shown.map((item) => (
          <li
            key={item.key}
            className="grid grid-cols-[minmax(0,10rem)_1fr_auto] items-center gap-3 text-sm sm:grid-cols-[minmax(0,14rem)_1fr_auto]"
          >
            <span className="leading-snug text-foreground">{item.label}</span>
            <span className="h-2 rounded-full bg-surface-muted">
              <span
                className="block h-2 rounded-full bg-brand-green"
                style={{ width: `${(item.count / max) * 100}%` }}
              />
            </span>
            <span className="tabular w-10 text-right font-mono text-xs text-muted-foreground">
              {formatNumber(item.count)}
            </span>
          </li>
        ))}
      </ul>
      {items.length > limit && (
        <button
          type="button"
          onClick={() => setExpanded((value) => !value)}
          className="mt-3 rounded-control text-sm font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
        >
          {expanded ? "Kevesebb" : `Mind (${items.length})`}
        </button>
      )}
    </div>
  );
}

export function RecordBadge({ record }: { record: string }) {
  const color = RECORD_BADGE_COLORS[record] ?? RECORD_BADGE_COLORS.CR;
  return (
    <span
      className={`ml-2 rounded px-1.5 py-0.5 font-mono text-[10px] font-bold ${color}`}
    >
      {record}
    </span>
  );
}
