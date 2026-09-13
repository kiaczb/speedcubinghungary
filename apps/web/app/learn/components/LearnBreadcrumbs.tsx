import Link from "next/link";
import { FaChevronRight } from "react-icons/fa";

export type Crumb = {
  label: string;
  /** Az utolsó (aktuális) morzsán nincs href — az nem link, hanem az itt-vagy. */
  href?: string;
};

type Props = {
  items: Crumb[];
  className?: string;
};

export default function LearnBreadcrumbs({ items, className = "" }: Props) {
  return (
    <nav aria-label="Morzsamenü" className={className}>
      <ol className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;

          return (
            <li key={`${item.label}-${index}`} className="flex items-center gap-2">
              {index > 0 && (
                <FaChevronRight
                  aria-hidden
                  className="h-2.5 w-2.5 shrink-0 text-subtle-foreground"
                />
              )}

              {item.href && !isLast ? (
                <Link
                  href={item.href}
                  className="rounded-sm transition-colors duration-200 ease-smooth hover:text-foreground"
                >
                  {item.label}
                </Link>
              ) : (
                <span
                  aria-current={isLast ? "page" : undefined}
                  className={isLast ? "font-medium text-foreground" : undefined}
                >
                  {item.label}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
