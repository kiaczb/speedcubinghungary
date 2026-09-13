import Link from "next/link";
import { FaArrowRight } from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import type { LearnEvent } from "../types";

type Props = {
  event: LearnEvent;
};

export default function EventCard({ event }: Props) {
  return (
    <Link
      href={`/learn/events/${event.id}`}
      className="group block rounded-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
    >
      <Card interactive padding="lg" className="flex h-full flex-col gap-3">
        <div className="flex items-start justify-between gap-3">
          <h2 className="text-lg font-semibold text-foreground">{event.name}</h2>
          <FaArrowRight
            aria-hidden
            className="mt-1.5 h-3.5 w-3.5 shrink-0 text-subtle-foreground transition-transform duration-200 ease-smooth group-hover:translate-x-0.5 group-hover:text-foreground"
          />
        </div>

        {event.description && (
          <p className="text-sm leading-relaxed text-muted-foreground">
            {event.description}
          </p>
        )}

        <p className="tabular mt-auto font-mono text-xs text-subtle-foreground">
          {event.caseCount} eset
        </p>
      </Card>
    </Link>
  );
}
