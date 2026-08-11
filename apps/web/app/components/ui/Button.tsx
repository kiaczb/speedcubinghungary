import Link from "next/link";
import React, { type AnchorHTMLAttributes, type ButtonHTMLAttributes } from "react";

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "inverse";
type ButtonSize = "sm" | "md" | "lg";

const BASE =
  "inline-flex items-center justify-center gap-2 rounded-control font-semibold " +
  "transition-[background-color,border-color,color,box-shadow,transform] duration-200 ease-smooth " +
  "focus-visible:outline-2 focus-visible:outline-offset-2 " +
  "disabled:pointer-events-none disabled:opacity-50 " +
  "active:translate-y-px";

const VARIANTS: Record<ButtonVariant, string> = {
  primary:
    "bg-foreground text-white shadow-card hover:bg-ink-hover hover:shadow-card-hover focus-visible:outline-foreground",
  secondary:
    "border border-line-strong text-foreground hover:border-foreground hover:bg-surface-muted focus-visible:outline-foreground",
  ghost:
    "text-muted-foreground hover:bg-surface-muted hover:text-foreground focus-visible:outline-foreground",
  danger:
    "bg-brand-red text-white shadow-card hover:bg-brand-red-strong hover:shadow-card-hover focus-visible:outline-brand-red",
  inverse:
    "bg-white text-foreground shadow-card hover:bg-surface-muted hover:shadow-card-hover focus-visible:outline-white",
};

const SIZES: Record<ButtonSize, string> = {
  sm: "px-3 py-1.5 text-sm",
  md: "px-5 py-2.5 text-sm",
  lg: "px-6 py-3.5 text-base",
};

type CommonProps = {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: React.ReactNode;
  /** Hoverre finoman jobbra csúszik. */
  iconRight?: React.ReactNode;
  fullWidth?: boolean;
  className?: string;
  children?: React.ReactNode;
};

type ButtonAsButton = CommonProps &
  Omit<ButtonHTMLAttributes<HTMLButtonElement>, "className" | "children"> & {
    href?: undefined;
  };

type ButtonAsLink = CommonProps &
  Omit<AnchorHTMLAttributes<HTMLAnchorElement>, "className" | "children" | "href"> & {
    href: string;
    /** Külső link: új fülön nyílik, `rel="noopener noreferrer"`-rel. */
    external?: boolean;
  };

type ButtonProps = ButtonAsButton | ButtonAsLink;

export default function Button(props: ButtonProps) {
  const {
    variant = "primary",
    size = "md",
    icon,
    iconRight,
    fullWidth,
    className = "",
    children,
    ...rest
  } = props;

  const classes = [
    BASE,
    VARIANTS[variant],
    SIZES[size],
    fullWidth ? "w-full" : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  const content = (
    <>
      {icon}
      {children}
      {iconRight && (
        <span className="transition-transform duration-200 ease-smooth group-hover:translate-x-0.5">
          {iconRight}
        </span>
      )}
    </>
  );

  // A `group` kell ahhoz, hogy a jobb oldali ikon a gomb hoverére reagáljon.
  const withGroup = `group ${classes}`;

  if (props.href !== undefined) {
    const { href, external, ...anchorRest } =
      rest as AnchorHTMLAttributes<HTMLAnchorElement> & {
        href: string;
        external?: boolean;
      };

    if (external) {
      return (
        <a
          {...anchorRest}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className={withGroup}
        >
          {content}
        </a>
      );
    }

    return (
      <Link {...anchorRest} href={href} className={withGroup}>
        {content}
      </Link>
    );
  }

  return (
    <button
      {...(rest as ButtonHTMLAttributes<HTMLButtonElement>)}
      className={withGroup}
    >
      {content}
    </button>
  );
}
