import type { Metadata } from "next";
import React from "react";
import PageHeader from "../components/ui/PageHeader";
import Card from "../components/Cards/Card";
import ContactForm from "./components/ContactForm";
import { CONTACT_ITEMS } from "../lib/site";

export const metadata: Metadata = {
  title: "Kapcsolat",
  description:
    "Vedd fel velünk a kapcsolatot e-mailben vagy a közösségi oldalunkon.",
};

const Contact = () => {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <PageHeader
          title="Kapcsolat"
          description="Kérdésed van a kockázással kapcsolatban, versenyt szerveznél, vagy részt vennél egy eseményen? Keress minket bátran."
        />

        <div className="mt-10 grid grid-cols-1 gap-10 lg:grid-cols-[1fr_1.2fr] lg:gap-14">
          <div>
            <h2 className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-subtle-foreground">
              Elérhetőségeink
            </h2>

            <ul className="mt-5 space-y-2">
              {CONTACT_ITEMS.map(({ icon: Icon, label, href }) => (
                <li key={label}>
                  <a
                    href={href}
                    className="group flex items-center gap-3 rounded-control border border-line bg-surface px-4 py-3 text-sm transition-colors duration-200 ease-smooth hover:border-line-strong hover:bg-surface-muted"
                  >
                    <Icon className="h-4 w-4 shrink-0 text-subtle-foreground transition-colors duration-200 ease-smooth group-hover:text-foreground" />
                    <span className="text-foreground">{label}</span>
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <Card padding="lg">
            <ContactForm />
          </Card>
        </div>
      </div>
    </div>
  );
};

export default Contact;
