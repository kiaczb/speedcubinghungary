import Link from "next/link";
import { FaChevronRight } from "react-icons/fa";
import { ALGORITHMS_HREF, LEARN_HREF, levelHref } from "../lib/nodes";
import type { LearnNode } from "../types";

type Props = {
  ancestors: LearnNode[];
  node: LearnNode | null;
};

type Crumb = {
  label: string;
  href?: string;
};

export default function LearnBreadcrumbs({ ancestors, node }: Props) {
  const crumbs: Crumb[] = [
    { label: "Tanulás", href: LEARN_HREF },
    { label: "Algoritmusok", href: node ? ALGORITHMS_HREF : undefined },
    ...ancestors.map((a) => ({ label: a.title, href: levelHref(a.path) })),
    ...(node ? [{ label: node.title }] : []),
  ];

  return (
    <nav aria-label="Morzsamenü">
      <ol className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
        {crumbs.map((crumb, index) => (
          <li key={`${index}-${crumb.label}`} className="flex items-center gap-2">
            {index > 0 && (
              <FaChevronRight
                className="h-2.5 w-2.5 text-subtle-foreground"
                aria-hidden
              />
            )}
            {crumb.href ? (
              <Link
                href={crumb.href}
                className="font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
              >
                {crumb.label}
              </Link>
            ) : (
              <span aria-current="page" className="font-medium text-foreground">
                {crumb.label}
              </span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
