import Link from "next/link";
import { FaArrowRight, FaVideo } from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import type { LearnAlgorithm } from "../types";

type Props = {
  algorithm: LearnAlgorithm;
};

export default function AlgorithmListItem({ algorithm }: Props) {
  return (
    <Link
      href={`/learn/algorithms/${algorithm.id}`}
      className="group block rounded-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
    >
      <Card interactive padding="lg" className="flex items-center gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
            <h2 className="font-semibold text-foreground">{algorithm.name}</h2>
            {algorithm.videoUrl && (
              <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
                <FaVideo aria-hidden className="h-3 w-3" />
                Videóval
              </span>
            )}
          </div>

          {/* Az algoritmus maga a lényeg — a listában is olvasható legyen,
              ne csak a tanulónézetben. */}
          <p className="tabular mt-1.5 break-words font-mono text-sm text-muted-foreground">
            {algorithm.notation}
          </p>
        </div>

        <FaArrowRight
          aria-hidden
          className="h-3.5 w-3.5 shrink-0 text-subtle-foreground transition-transform duration-200 ease-smooth group-hover:translate-x-0.5 group-hover:text-foreground"
        />
      </Card>
    </Link>
  );
}
