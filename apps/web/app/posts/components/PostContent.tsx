import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Components } from "react-markdown";
import { RECORD_BADGE_COLORS } from "../lib/recordBadges";

/**
 * BIZTONSÁG: nincs `rehype-raw`, ezért a react-markdown a forrásban lévő nyers
 * HTML-t NEM rendereli (escape-eli). Így XSS-t nem lehet beinjektálni még akkor
 * sem, ha egy admin fiók kompromittálódik. Ne adj hozzá rehype-raw-t.
 */

const components: Components = {
  h1: ({ children }) => (
    <h1 className="mb-4 mt-10 text-2xl font-bold first:mt-0">{children}</h1>
  ),
  h2: ({ children }) => (
    <h2 className="mb-3 mt-10 border-b border-line pb-2 text-xl font-bold first:mt-0">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="mb-2 mt-7 text-lg font-semibold first:mt-0">{children}</h3>
  ),
  p: ({ children }) => (
    <p className="my-4 leading-[1.75] text-muted-foreground">{children}</p>
  ),
  a: ({ href, children }) => (
    <a
      href={href}
      className="font-medium text-foreground underline decoration-line-strong underline-offset-2 transition-colors duration-200 ease-smooth hover:decoration-foreground"
    >
      {children}
    </a>
  ),
  ul: ({ children }) => (
    <ul className="my-4 list-disc space-y-1.5 pl-6 text-muted-foreground marker:text-line-strong">
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="my-4 list-decimal space-y-1.5 pl-6 text-muted-foreground marker:text-subtle-foreground">
      {children}
    </ol>
  ),
  blockquote: ({ children }) => (
    <blockquote className="my-6 border-l-2 border-brand-red pl-5 italic text-muted-foreground">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-10 border-line" />,
  strong: ({ children }) => (
    <strong className="font-semibold text-foreground">{children}</strong>
  ),
  // Inline kód: ha a tartalma egy ismert rekord-jelölés, színes chippé alakítjuk.
  code: ({ children }) => {
    const text = String(children);
    const badge = RECORD_BADGE_COLORS[text];
    if (badge) {
      return (
        <span
          className={`mx-0.5 inline-block rounded px-1.5 py-0.5 font-mono text-xs font-bold ${badge}`}
        >
          {text}
        </span>
      );
    }
    return (
      <code className="rounded bg-surface-muted px-1.5 py-0.5 font-mono text-[0.85em] text-foreground">
        {children}
      </code>
    );
  },
  // Vízszintesen görgethető konténer, hogy mobilon ne törje szét a layoutot.
  table: ({ children }) => (
    <div className="my-6 overflow-x-auto rounded-card border border-line">
      <table className="w-full border-collapse text-sm">{children}</table>
    </div>
  ),
  thead: ({ children }) => (
    <thead className="border-b border-line bg-surface-muted text-left">
      {children}
    </thead>
  ),
  th: ({ children }) => (
    <th className="px-4 py-2.5 font-semibold text-foreground">{children}</th>
  ),
  td: ({ children }) => (
    <td className="tabular border-b border-line px-4 py-2.5 text-muted-foreground last:border-0">
      {children}
    </td>
  ),
};

type Props = {
  content: string;
  className?: string;
};

export default function PostContent({ content, className }: Props) {
  return (
    <div className={className}>
      <Markdown remarkPlugins={[remarkGfm]} components={components}>
        {content}
      </Markdown>
    </div>
  );
}
