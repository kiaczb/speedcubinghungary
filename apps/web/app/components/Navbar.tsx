"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import React, { useEffect, useState } from "react";
import SpeedcubingFullAnimatedSVG from "./svgs/SpeedcubingFullAnimatedSVG";
import { NAV_LINKS } from "../lib/site";

type NavbarProps = {
  action?: React.ReactNode;
};

const Navbar = ({ action }: NavbarProps) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  const pathname = usePathname();

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 0);
    };

    handleScroll();
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const isActive = (href: string) =>
    pathname === href || pathname.startsWith(`${href}/`);

  return (
    <nav
      className={`sticky top-0 z-50 w-full border-b bg-white transition-[border-color,box-shadow] duration-300 ease-smooth dark:bg-canvas ${
        isScrolled ? "border-line shadow-card" : "border-transparent"
      }`}
    >
      <div className="mx-auto w-full max-w-6xl px-5 sm:px-8 lg:px-12">
        <div className="flex h-16 items-center justify-between gap-6">
          <Link
            href="/"
            aria-label="Speedcubing Hungary — főoldal"
            className="shrink-0 rounded-control"
          >
            <SpeedcubingFullAnimatedSVG animated={false} maxWidth={196} />
          </Link>

          <div className="hidden items-center gap-1 md:flex">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                aria-current={isActive(link.href) ? "page" : undefined}
                className={`relative rounded-control px-3 py-2 text-sm font-medium transition-colors duration-200 ease-smooth ${
                  isActive(link.href)
                    ? "text-foreground"
                    : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
                }`}
              >
                {link.label}
                <span
                  className={`absolute inset-x-3 -bottom-px h-0.5 rounded-full bg-brand-red transition-opacity duration-200 ease-smooth ${
                    isActive(link.href) ? "opacity-100" : "opacity-0"
                  }`}
                />
              </Link>
            ))}
            {action && <div className="ml-3">{action}</div>}
          </div>

          <button
            type="button"
            onClick={() => setIsMenuOpen((open) => !open)}
            aria-expanded={isMenuOpen}
            aria-controls="mobile-menu"
            aria-label={isMenuOpen ? "Menü bezárása" : "Menü megnyitása"}
            className="inline-flex h-10 w-10 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground md:hidden"
          >
            <svg
              className="h-5 w-5"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d={isMenuOpen ? "M6 18L18 6M6 6l12 12" : "M4 6h16M4 12h16M4 18h16"}
              />
            </svg>
          </button>
        </div>
      </div>

      <div
        id="mobile-menu"
        inert={!isMenuOpen}
        className={`overflow-hidden border-t border-line bg-white transition-[max-height,opacity] duration-300 ease-smooth dark:bg-surface md:hidden ${
          isMenuOpen ? "max-h-96 opacity-100" : "max-h-0 border-transparent opacity-0"
        }`}
      >
        <div className="flex flex-col gap-1 px-5 py-4 sm:px-8">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              onClick={() => setIsMenuOpen(false)}
              aria-current={isActive(link.href) ? "page" : undefined}
              className={`rounded-control px-3 py-2.5 text-base font-medium transition-colors duration-200 ease-smooth ${
                isActive(link.href)
                  ? "bg-surface-muted text-foreground"
                  : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
              }`}
            >
              {link.label}
            </Link>
          ))}
          {action && <div className="mt-3 border-t border-line pt-4">{action}</div>}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
