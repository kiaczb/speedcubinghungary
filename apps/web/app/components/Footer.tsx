import Link from "next/link";
import React from "react";
import { CONTACT_ITEMS, NAV_LINKS, SITE_DESCRIPTION } from "../lib/site";

const Footer = () => {
  return (
    <footer className="border-t border-line bg-surface-inverse text-white">
      <div className="mx-auto w-full max-w-6xl px-5 py-10 sm:px-8 lg:px-12">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-[1.5fr_1fr_1fr]">
          <div>
            <h2 className="text-lg font-bold tracking-tight">
              Speedcubing Hungary
            </h2>
            <p className="mt-3 max-w-sm text-sm leading-relaxed text-white/60">
              {SITE_DESCRIPTION}
            </p>
          </div>

          <div>
            <h2 className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-white/40">
              Oldalak
            </h2>
            <ul className="mt-4 space-y-2.5">
              {NAV_LINKS.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-sm text-white/70 transition-colors duration-200 ease-smooth hover:text-white"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h2 className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-white/40">
              Elérhetőség
            </h2>
            <ul className="mt-4 space-y-2.5">
              {CONTACT_ITEMS.map(({ icon: Icon, label, href }) => (
                <li key={label}>
                  <a
                    href={href}
                    className="group inline-flex items-center gap-2.5 text-sm text-white/70 transition-colors duration-200 ease-smooth hover:text-white"
                  >
                    <Icon className="h-4 w-4 shrink-0 text-white/40 transition-colors duration-200 ease-smooth group-hover:text-white/70" />
                    <span>{label}</span>
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="mt-8 border-t border-white/10 pt-5">
          <p className="font-mono text-xs text-white/40">
            © {new Date().getFullYear()} Speedcubing Hungary. Minden jog
            fenntartva.
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
