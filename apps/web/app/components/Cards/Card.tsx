import React from "react";

/**
 * Az oldal egyetlen kártya-alapja.
 *
 * Korábban minden felület a saját változatát rajzolta meg — máshol volt az
 * árnyék, a sugár és a keret —, ezért a kártyák egymás mellett sosem néztek ki
 * egy rendszer részének. A `variant` most ezt zárja egységbe; a `className`
 * továbbra is felülírhat bármit.
 *
 * A korábbi `imageUrl` / `imageClass` propok törölve lettek: egyetlen hívó sem
 * használta őket (a képet mindenhol maga a hívó rendereli).
 */

type CardVariant = "elevated" | "outlined" | "inverse";
type CardPadding = "none" | "sm" | "md" | "lg";

const VARIANTS: Record<CardVariant, string> = {
  // Alapértelmezett: fehér lap, finom kerettel és lágy árnyékkal.
  elevated: "bg-surface border border-line shadow-card",
  // Laposabb változat listákhoz, ahol sok kártya van egymás alatt.
  outlined: "bg-surface border border-line",
  // Sötét kártya világos szöveggel (hero, kiemelt blokkok).
  inverse: "bg-surface-inverse text-white border border-white/10 shadow-panel",
};

const PADDINGS: Record<CardPadding, string> = {
  none: "",
  sm: "p-4",
  md: "p-5",
  lg: "p-6",
};

type CardProps = {
  variant?: CardVariant;
  padding?: CardPadding;
  /** Hoverre megemeli a kártyát — csak kattintható kártyákon használd. */
  interactive?: boolean;
  className?: string;
  children: React.ReactNode;
};

const Card = ({
  variant = "elevated",
  padding = "none",
  interactive = false,
  className = "",
  children,
}: CardProps) => {
  const classes = [
    "rounded-card w-full",
    VARIANTS[variant],
    PADDINGS[padding],
    interactive
      ? "transition-[box-shadow,transform,border-color] duration-200 ease-smooth hover:-translate-y-0.5 hover:shadow-card-hover hover:border-line-strong"
      : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return <div className={classes}>{children}</div>;
};

export default Card;
