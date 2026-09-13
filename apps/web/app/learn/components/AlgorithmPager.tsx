import Link from "next/link";
import { FaChevronLeft, FaChevronRight } from "react-icons/fa";
import type { LearnAlgorithm } from "../types";

type Props = {
  siblings: LearnAlgorithm[];
  currentId: string;
  caseName: string;
};

// Tanulás közben a következő algoritmusra lépés a leggyakoribb mozdulat —
// ehhez ne kelljen visszanavigálni az eset oldalára. A lista az eset ÖSSZES
// algoritmusát tartalmazza, ezért a szomszédok kérés nélkül adódnak.
export default function AlgorithmPager({
  siblings,
  currentId,
  caseName,
}: Props) {
  const index = siblings.findIndex((item) => item.id === currentId);
  if (index === -1 || siblings.length < 2) return null;

  const previous = index > 0 ? siblings[index - 1] : null;
  const next = index < siblings.length - 1 ? siblings[index + 1] : null;

  return (
    <section
      aria-label="Az eset többi algoritmusa"
      className="mt-10 border-t border-line pt-6"
    >
      <div className="flex flex-wrap items-center justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          <span className="tabular font-mono">
            {index + 1}/{siblings.length}
          </span>{" "}
          — {caseName}
        </p>

        <div className="flex items-center gap-2">
          <PagerLink algorithm={previous} direction="previous" />
          <PagerLink algorithm={next} direction="next" />
        </div>
      </div>
    </section>
  );
}

type PagerLinkProps = {
  algorithm: LearnAlgorithm | null;
  direction: "previous" | "next";
};

function PagerLink({ algorithm, direction }: PagerLinkProps) {
  const isPrevious = direction === "previous";
  const label = isPrevious ? "Előző" : "Következő";

  // A szélen nem tüntetjük el a gombot, csak letiltjuk — így a két gomb helye
  // nem ugrál lapozás közben.
  if (!algorithm) {
    return (
      <span
        aria-disabled
        className="inline-flex cursor-not-allowed items-center gap-2 rounded-control border border-line px-3 py-1.5 text-sm font-medium text-subtle-foreground opacity-60"
      >
        {isPrevious && <FaChevronLeft aria-hidden className="h-3 w-3" />}
        {label}
        {!isPrevious && <FaChevronRight aria-hidden className="h-3 w-3" />}
      </span>
    );
  }

  return (
    <Link
      href={`/learn/algorithms/${algorithm.id}`}
      title={algorithm.name}
      className="inline-flex max-w-[12rem] items-center gap-2 rounded-control border border-line-strong px-3 py-1.5 text-sm font-medium text-foreground transition-[background-color,border-color] duration-200 ease-smooth hover:border-foreground hover:bg-surface-muted"
    >
      {isPrevious && <FaChevronLeft aria-hidden className="h-3 w-3 shrink-0" />}
      <span className="truncate">{algorithm.name}</span>
      {!isPrevious && <FaChevronRight aria-hidden className="h-3 w-3 shrink-0" />}
    </Link>
  );
}
