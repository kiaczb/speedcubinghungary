import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Components } from "react-markdown";

/**
 * A poszt markdown tartalmának renderelése — stílussal.
 *
 * BIZTONSÁG: nincs `rehype-raw`, ezért a react-markdown a forrásban lévő nyers
 * HTML-t NEM rendereli (escape-eli). Így a content markdownból stílusozható, de
 * XSS-t nem lehet beinjektálni még akkor sem, ha egy admin fiók kompromittálódik
 * vagy a weekly-report generálja a tartalmat. Ne adj hozzá rehype-raw-t.
 *
 * A megjelenést a `components` map adja (Tailwind), nem maga a markdown — a
 * struktúrát (címek, táblázat, idézet, badge) a markdown írja le, a kinézetet mi.
 */

// Rekord-badge-ek (NR/CR/WR) színezése, hasonlóan az e-mailes riporthoz.
// A markdownban ezek inline kódként jönnek (pl. `NR`), így itt tudjuk chippé
// alakítani őket anélkül, hogy nyers HTML kellene a forrásba.
const badgeColors: Record<string, string> = {
  WR: "bg-yellow-400 text-yellow-950",
  CR: "bg-blue-500 text-white",
  NR: "bg-green-600 text-white",
  PR: "bg-gray-400 text-gray-950",
};

const components: Components = {
  h1: ({ children }) => (
    <h1 className="mb-4 mt-8 text-3xl font-bold first:mt-0">{children}</h1>
  ),
  h2: ({ children }) => (
    <h2 className="mb-3 mt-8 border-b border-gray-200 pb-1 text-2xl font-bold first:mt-0 dark:border-gray-700">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="mb-2 mt-6 text-xl font-semibold first:mt-0">{children}</h3>
  ),
  p: ({ children }) => <p className="my-3 leading-relaxed">{children}</p>,
  a: ({ href, children }) => (
    <a
      href={href}
      className="text-blue-600 underline hover:text-blue-800 dark:text-blue-400"
    >
      {children}
    </a>
  ),
  ul: ({ children }) => (
    <ul className="my-3 list-disc space-y-1 pl-6">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="my-3 list-decimal space-y-1 pl-6">{children}</ol>
  ),
  blockquote: ({ children }) => (
    <blockquote className="my-4 border-l-4 border-gray-300 pl-4 italic text-gray-600 dark:border-gray-600 dark:text-gray-400">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-6 border-gray-200 dark:border-gray-700" />,
  strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
  // Inline kód: ha a tartalma egy ismert rekord-badge, színes chippé alakítjuk.
  code: ({ children }) => {
    const text = String(children);
    const badge = badgeColors[text];
    if (badge) {
      return (
        <span
          className={`mx-0.5 inline-block rounded px-1.5 py-0.5 text-xs font-bold ${badge}`}
        >
          {text}
        </span>
      );
    }
    return (
      <code className="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-sm dark:bg-gray-800">
        {children}
      </code>
    );
  },
  // A táblázat vízszintesen görgethető konténerben, hogy mobilon ne törje szét a layoutot.
  table: ({ children }) => (
    <div className="my-4 overflow-x-auto">
      <table className="w-full border-collapse text-sm">{children}</table>
    </div>
  ),
  thead: ({ children }) => (
    <thead className="border-b-2 border-gray-300 text-left dark:border-gray-600">
      {children}
    </thead>
  ),
  th: ({ children }) => <th className="px-3 py-2 font-semibold">{children}</th>,
  td: ({ children }) => (
    <td className="border-b border-gray-100 px-3 py-2 dark:border-gray-800">
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
