import React from "react";

/**
 * Kategória-címke a posztokhoz.
 *
 * A `color` a hívótól jön (lásd `badgeMap` az app/posts/types.ts-ben), és most
 * teljes szín-hármast ad: háttér + szöveg + gyűrű. A korábbi telített, fehér
 * szövegű változat a sárga "Közlemény" címkén olvashatatlan volt.
 */

type BadgeProps = {
  label: string;
  /** Tailwind osztályok a tónushoz (háttér + szöveg + ring). */
  color: string;
  className?: string;
};

const Badge: React.FC<BadgeProps> = ({ label, color, className = "" }) => {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold uppercase tracking-wide ring-1 ring-inset ${color} ${className}`}
    >
      {label}
    </span>
  );
};

export default Badge;
