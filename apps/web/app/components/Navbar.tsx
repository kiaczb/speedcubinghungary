"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import React, { useEffect, useRef, useState } from "react";
import { FaChevronDown } from "react-icons/fa";
import SpeedcubingFullAnimatedSVG from "./svgs/SpeedcubingFullAnimatedSVG";
import { NAV_ITEMS, NavGroup, NavLink, isNavGroup } from "../lib/site";

type NavbarProps = {
  action?: React.ReactNode;
};

const Navbar = ({ action }: NavbarProps) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  const [openGroup, setOpenGroup] = useState<string | null>(null);
  const [openMobileGroup, setOpenMobileGroup] = useState<string | null>(null);
  const desktopNavRef = useRef<HTMLDivElement>(null);
  const lastPointerType = useRef<string | null>(null);
  const pathname = usePathname();

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 0);
    };

    handleScroll();
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    if (!openGroup) return;

    const handlePointerDown = (event: PointerEvent) => {
      if (!desktopNavRef.current?.contains(event.target as Node)) {
        setOpenGroup(null);
      }
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpenGroup(null);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [openGroup]);

  const isActive = (href: string) =>
    pathname === href || pathname.startsWith(`${href}/`);

  const isGroupActive = (group: NavGroup) =>
    group.links.some((link) => isActive(link.href));

  const topLevelClass = (active: boolean) =>
    `relative inline-flex items-center gap-1.5 rounded-control px-3 py-2 text-sm font-medium transition-colors duration-200 ease-smooth ${
      active
        ? "text-foreground"
        : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
    }`;

  const activeBar = (active: boolean) => (
    <span
      className={`absolute inset-x-3 -bottom-px h-0.5 rounded-full bg-brand-red transition-opacity duration-200 ease-smooth ${
        active ? "opacity-100" : "opacity-0"
      }`}
    />
  );

  const renderDesktopLink = (link: NavLink) => (
    <Link
      key={link.href}
      href={link.href}
      aria-current={isActive(link.href) ? "page" : undefined}
      className={topLevelClass(isActive(link.href))}
    >
      {link.label}
      {activeBar(isActive(link.href))}
    </Link>
  );

  const renderDesktopGroup = (group: NavGroup) => {
    const isOpen = openGroup === group.label;
    const panelId = `nav-group-${group.label}`;
    const active = isGroupActive(group);

    const close = () =>
      setOpenGroup((current) => (current === group.label ? null : current));

    return (
      <div
        key={group.label}
        className="relative"
        onPointerEnter={(event) => {
          if (event.pointerType === "mouse") setOpenGroup(group.label);
        }}
        onPointerLeave={(event) => {
          if (event.pointerType === "mouse") close();
        }}
        onBlur={(event) => {
          if (!event.currentTarget.contains(event.relatedTarget as Node)) {
            close();
          }
        }}
      >
        <button
          type="button"
          aria-expanded={isOpen}
          aria-controls={panelId}
          onPointerDown={(event) => {
            lastPointerType.current = event.pointerType;
          }}
          onClick={() => {
            const fromMouse = lastPointerType.current === "mouse";
            lastPointerType.current = null;
            setOpenGroup(fromMouse || !isOpen ? group.label : null);
          }}
          className={`${topLevelClass(active)} ${isOpen ? "bg-surface-muted text-foreground" : ""}`}
        >
          {group.label}
          <FaChevronDown
            className={`h-2.5 w-2.5 text-subtle-foreground transition-transform duration-200 ease-smooth ${
              isOpen ? "rotate-180" : ""
            }`}
            aria-hidden="true"
          />
          {activeBar(active)}
        </button>

        <div
          id={panelId}
          hidden={!isOpen}
          className="absolute left-0 top-full z-10 pt-2"
        >
          <ul className="flex min-w-44 flex-col gap-0.5 rounded-card border border-line bg-surface p-1.5 shadow-panel">
            {group.links.map(({ icon: Icon, ...link }) => (
              <li key={link.href}>
                <Link
                  href={link.href}
                  onClick={() => setOpenGroup(null)}
                  aria-current={isActive(link.href) ? "page" : undefined}
                  className={`group/item flex items-center gap-2.5 whitespace-nowrap rounded-control px-3 py-2 text-sm font-medium transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground ${
                    isActive(link.href)
                      ? "bg-surface-muted text-foreground"
                      : "text-muted-foreground"
                  }`}
                >
                  <Icon
                    aria-hidden="true"
                    className={`h-3.5 w-3.5 shrink-0 transition-colors duration-200 ease-smooth group-hover/item:text-brand-red ${
                      isActive(link.href) ? "text-brand-red" : "text-subtle-foreground"
                    }`}
                  />
                  {link.label}
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </div>
    );
  };

  const renderMobileLink = ({ icon: Icon, ...link }: NavLink) => (
    <Link
      key={link.href}
      href={link.href}
      onClick={() => setIsMenuOpen(false)}
      aria-current={isActive(link.href) ? "page" : undefined}
      className={`flex items-center gap-3 rounded-control px-3 py-2.5 text-base font-medium transition-colors duration-200 ease-smooth ${
        isActive(link.href)
          ? "bg-surface-muted text-foreground"
          : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
      }`}
    >
      <Icon
        aria-hidden="true"
        className={`h-4 w-4 shrink-0 ${
          isActive(link.href) ? "text-brand-red" : "text-subtle-foreground"
        }`}
      />
      {link.label}
    </Link>
  );

  const renderMobileGroup = (group: NavGroup) => {
    const isOpen = openMobileGroup === group.label;
    const panelId = `mobile-nav-group-${group.label}`;
    const active = isGroupActive(group);

    return (
      <div key={group.label} className="flex flex-col">
        <button
          type="button"
          aria-expanded={isOpen}
          aria-controls={panelId}
          onClick={() => setOpenMobileGroup(isOpen ? null : group.label)}
          className={`flex items-center justify-between rounded-control px-3 py-2.5 text-base font-medium transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground ${
            active || isOpen ? "text-foreground" : "text-muted-foreground"
          }`}
        >
          {group.label}
          <FaChevronDown
            aria-hidden="true"
            className={`h-3 w-3 text-subtle-foreground transition-transform duration-200 ease-smooth ${
              isOpen ? "rotate-180" : ""
            }`}
          />
        </button>

        <div
          id={panelId}
          inert={!isOpen}
          className={`grid transition-[grid-template-rows,opacity] duration-200 ease-smooth ${
            isOpen ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0"
          }`}
        >
          <div className="overflow-hidden">
            <div className="ml-3 flex flex-col gap-1 border-l border-line py-1 pl-3">
              {group.links.map(renderMobileLink)}
            </div>
          </div>
        </div>
      </div>
    );
  };

  return (
    <nav
      className={`sticky top-0 z-50 w-full border-b bg-white transition-[border-color,box-shadow] duration-300 ease-smooth ${
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

          <div ref={desktopNavRef} className="hidden items-center gap-1 md:flex">
            {NAV_ITEMS.map((item) =>
              isNavGroup(item)
                ? renderDesktopGroup(item)
                : renderDesktopLink(item),
            )}
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
        className={`overflow-y-auto border-t border-line bg-white transition-[max-height,opacity] duration-300 ease-smooth md:hidden ${
          isMenuOpen
            ? "max-h-[calc(100vh-4rem)] opacity-100"
            : "max-h-0 border-transparent opacity-0"
        }`}
      >
        <div className="flex flex-col gap-1 px-5 py-4 sm:px-8">
          {NAV_ITEMS.map((item) =>
            isNavGroup(item) ? renderMobileGroup(item) : renderMobileLink(item),
          )}
          {action && <div className="mt-3 border-t border-line pt-4">{action}</div>}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
