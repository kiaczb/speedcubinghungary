import Link from "next/link";
import "@cubing/icons";
import { FaArrowRight } from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import { contentSummary, levelHref } from "../lib/nodes";
import type { LearnNode } from "../types";

type Props = {
  node: LearnNode;
  showMedia?: boolean;
  actions?: React.ReactNode;
};

const LINK_CLASS =
  "after:absolute after:inset-0 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-foreground";

export default function NodeCard({ node, showMedia = false, actions }: Props) {
  if (showMedia) {
    return (
      <Card padding="md" className="relative flex h-full flex-col items-center text-center transition-[border-color,box-shadow] duration-200 ease-smooth hover:border-line-strong hover:shadow-card-hover">
        <div className="flex h-24 w-full items-center justify-center sm:h-28" aria-hidden>
          {node.imageUrl ? (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={node.imageUrl}
              alt=""
              loading="lazy"
              className="h-full w-auto max-w-full object-contain"
            />
          ) : node.icon ? (
            <span className={`cubing-icon ${node.icon} text-6xl leading-none text-foreground sm:text-7xl`} />
          ) : (
            <span className="font-mono text-5xl font-semibold text-subtle-foreground">
              {node.title.charAt(0).toUpperCase()}
            </span>
          )}
        </div>

        <h2 className="mt-3 w-full break-words text-sm font-semibold leading-snug text-foreground sm:text-base">
          <Link href={levelHref(node.path)} className={LINK_CLASS}>
            {node.title}
          </Link>
        </h2>

        <p className="tabular mt-1 font-mono text-xs text-subtle-foreground">
          {contentSummary(node)}
        </p>

        {actions && <div className="mt-auto w-full pt-2">{actions}</div>}
      </Card>
    );
  }

  return (
    <Card interactive padding="md" className="group relative flex h-full flex-col gap-2">
      <div className="flex items-start justify-between gap-3">
        <h2 className="min-w-0 break-words text-base font-semibold leading-snug text-foreground">
          <Link href={levelHref(node.path)} className={LINK_CLASS}>
            {node.title}
          </Link>
        </h2>
        <FaArrowRight
          className="mt-1 h-3 w-3 shrink-0 text-subtle-foreground transition-[transform,color] duration-200 ease-smooth group-hover:translate-x-0.5 group-hover:text-foreground"
          aria-hidden
        />
      </div>

      {node.description && (
        <p className="line-clamp-2 text-sm leading-relaxed text-muted-foreground">
          {node.description}
        </p>
      )}

      <p className="tabular mt-auto pt-1 font-mono text-xs text-subtle-foreground">
        {contentSummary(node)}
      </p>

      {actions}
    </Card>
  );
}
