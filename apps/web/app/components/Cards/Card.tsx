import React from "react";

type CardPadding = "none" | "md" | "lg";

const PADDINGS: Record<CardPadding, string> = {
  none: "",
  md: "p-5",
  lg: "p-6",
};

type CardProps = {
  padding?: CardPadding;
  /** Hoverre megemeli a kártyát — csak kattintható kártyákon használd. */
  interactive?: boolean;
  className?: string;
  children: React.ReactNode;
};

const Card = ({
  padding = "none",
  interactive = false,
  className = "",
  children,
}: CardProps) => {
  const classes = [
    "rounded-card w-full bg-surface border border-line shadow-card",
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
