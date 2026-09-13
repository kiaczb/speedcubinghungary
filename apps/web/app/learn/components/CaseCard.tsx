import Link from "next/link";
import { FaArrowRight } from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import type { LearnCase } from "../types";

type Props = {
  learnCase: LearnCase;
};

export default function CaseCard({ learnCase }: Props) {
  return (
    <Link
      href={`/learn/cases/${learnCase.id}`}
      className="group block rounded-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
    >
      <Card interactive padding="lg" className="flex h-full flex-col gap-3">
        <div className="flex items-start justify-between gap-3">
          <h2 className="font-semibold text-foreground">{learnCase.name}</h2>
          <FaArrowRight
            aria-hidden
            className="mt-1 h-3.5 w-3.5 shrink-0 text-subtle-foreground transition-transform duration-200 ease-smooth group-hover:translate-x-0.5 group-hover:text-foreground"
          />
        </div>

        {learnCase.description && (
          <p className="text-sm leading-relaxed text-muted-foreground">
            {learnCase.description}
          </p>
        )}

        <p className="tabular mt-auto font-mono text-xs text-subtle-foreground">
          {learnCase.algorithmCount} algoritmus
        </p>
      </Card>
    </Link>
  );
}
