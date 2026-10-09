import type { LearnAlgorithm } from "../types";
import CopyNotationButton from "./CopyNotationButton";
import VideoPlaceholder from "./VideoPlaceholder";

type Props = {
  algorithm: LearnAlgorithm;
  index: number;
  caseTitle: string;
  actions?: React.ReactNode;
};

export default function AlgorithmItem({ algorithm, index, caseTitle, actions }: Props) {
  return (
    <li className="flex flex-col gap-4 px-5 py-5 sm:flex-row sm:gap-6 sm:px-6">
      <div className="flex min-w-0 flex-1 gap-4">
        <span className="tabular mt-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-surface-muted font-mono text-xs text-muted-foreground">
          {index + 1}
        </span>

        <div className="min-w-0 flex-1">
          {algorithm.name && (
            <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-subtle-foreground">
              {algorithm.name}
            </p>
          )}

          <div className="flex items-start gap-2">
            <p className="min-w-0 flex-1 break-words font-mono text-lg leading-relaxed text-foreground [word-spacing:0.2em] sm:text-xl">
              {algorithm.notation}
            </p>
            <CopyNotationButton notation={algorithm.notation} />
          </div>

          {algorithm.notes && (
            <p className="mt-2 whitespace-pre-line text-sm leading-relaxed text-muted-foreground">
              {algorithm.notes}
            </p>
          )}

          {actions}
        </div>
      </div>

      <VideoPlaceholder
        compact
        label={`${caseTitle}, ${index + 1}. algoritmus`}
        className="sm:w-44 sm:shrink-0"
      />
    </li>
  );
}
