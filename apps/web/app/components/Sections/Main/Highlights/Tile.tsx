import Link from "next/link";
import React from "react";
import { FaChevronRight } from "react-icons/fa";
import Card from "../../../Cards/Card";

type TileProps = {
  href: string;
  title: string;
  meta?: React.ReactNode;
  cta: string;
  linkWholeCard?: boolean;
  className?: string;
  children: React.ReactNode;
};

export const Tile = ({
  href,
  title,
  meta,
  cta,
  linkWholeCard = true,
  className = "",
  children,
}: TileProps) => {
  const header = (
    <div className="flex items-baseline justify-between gap-3">
      <h3 className="text-lg font-semibold text-foreground">{title}</h3>
      {meta}
    </div>
  );

  const ctaContent = (
    <>
      {cta}
      <FaChevronRight className="h-2.5 w-2.5 text-subtle-foreground transition-transform duration-200 ease-smooth group-hover:translate-x-0.5 group-hover/cta:translate-x-0.5" />
    </>
  );

  if (!linkWholeCard) {
    return (
      <Card className={className}>
        <div className="flex h-full flex-col p-6">
          {header}
          <div className="mt-4 flex flex-1 flex-col">{children}</div>
          <Link
            href={href}
            className="group/cta mt-6 inline-flex items-center gap-1.5 self-start rounded-control text-sm font-medium text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
          >
            {ctaContent}
          </Link>
        </div>
      </Card>
    );
  }

  return (
    <Card interactive className={`group ${className}`}>
      <Link
        href={href}
        className="flex h-full flex-col rounded-card p-6 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
      >
        {header}
        <div className="mt-4 flex flex-1 flex-col">{children}</div>
        <span className="mt-6 inline-flex items-center gap-1.5 text-sm font-medium text-foreground">
          {ctaContent}
        </span>
      </Link>
    </Card>
  );
};

export const TileSkeleton = ({
  className = "",
  lines = 2,
}: {
  className?: string;
  lines?: number;
}) => (
  <Card className={className}>
    <div className="flex h-full animate-pulse flex-col p-6" aria-hidden="true">
      <div className="h-5 w-1/2 rounded-full bg-surface-muted" />
      {Array.from({ length: lines }, (_, index) => (
        <div
          key={index}
          className="mt-4 h-3 w-full rounded-full bg-surface-muted last:w-1/2"
        />
      ))}
    </div>
  </Card>
);

export const LiveIndicator = ({ label }: { label: string }) => (
  <span className="inline-flex items-center gap-2 rounded-full bg-brand-red/10 px-2.5 py-1 text-xs font-semibold text-brand-red">
    <span className="relative flex h-2 w-2">
      <span className="absolute inline-flex h-full w-full rounded-full bg-brand-red opacity-75 motion-safe:animate-ping" />
      <span className="relative inline-flex h-2 w-2 rounded-full bg-brand-red" />
    </span>
    {label}
  </span>
);
