"use client";

import { useLayoutEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
import {
  FaBold,
  FaCode,
  FaItalic,
  FaLink,
  FaListOl,
  FaListUl,
  FaQuoteLeft,
} from "react-icons/fa";

const PostContent = dynamic(() => import("./PostContent"), {
  ssr: false,
  loading: () => (
    <p className="text-sm text-muted-foreground">Előnézet betöltése…</p>
  ),
});

type Action =
  | { kind: "wrap"; before: string; after: string; sample: string }
  | { kind: "line"; prefix: string | ((index: number) => string); sample: string }
  | { kind: "link" };

type ToolbarItem = {
  label: string;
  title: string;
  icon: React.ReactNode;
  action: Action;
};

const TOOLBAR: ToolbarItem[] = [
  {
    label: "Félkövér",
    title: "Félkövér",
    icon: <FaBold className="h-3 w-3" />,
    action: { kind: "wrap", before: "**", after: "**", sample: "félkövér" },
  },
  {
    label: "Dőlt",
    title: "Dőlt",
    icon: <FaItalic className="h-3 w-3" />,
    action: { kind: "wrap", before: "*", after: "*", sample: "dőlt" },
  },
  {
    label: "Alcím",
    title: "Alcím (H2)",
    icon: <span className="text-xs font-bold">H2</span>,
    action: { kind: "line", prefix: "## ", sample: "Alcím" },
  },
  {
    label: "Kisebb alcím",
    title: "Kisebb alcím (H3)",
    icon: <span className="text-xs font-bold">H3</span>,
    action: { kind: "line", prefix: "### ", sample: "Alcím" },
  },
  {
    label: "Idézet",
    title: "Idézet",
    icon: <FaQuoteLeft className="h-3 w-3" />,
    action: { kind: "line", prefix: "> ", sample: "Idézet" },
  },
  {
    label: "Kód",
    title: "Kód",
    icon: <FaCode className="h-3 w-3" />,
    action: { kind: "wrap", before: "`", after: "`", sample: "kód" },
  },
  {
    label: "Felsorolás",
    title: "Felsorolás",
    icon: <FaListUl className="h-3 w-3" />,
    action: { kind: "line", prefix: "- ", sample: "Elem" },
  },
  {
    label: "Számozott lista",
    title: "Számozott lista",
    icon: <FaListOl className="h-3 w-3" />,
    action: { kind: "line", prefix: (i) => `${i + 1}. `, sample: "Elem" },
  },
  {
    label: "Link",
    title: "Link beszúrása",
    icon: <FaLink className="h-3 w-3" />,
    action: { kind: "link" },
  },
];

type Props = {
  id: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  rows?: number;
};

export default function MarkdownEditor({
  id,
  value,
  onChange,
  placeholder,
  rows = 12,
}: Props) {
  const [tab, setTab] = useState<"write" | "preview">("write");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const pendingSelection = useRef<[number, number] | null>(null);

  useLayoutEffect(() => {
    const selection = pendingSelection.current;
    const el = textareaRef.current;
    if (selection && el) {
      el.focus();
      el.setSelectionRange(selection[0], selection[1]);
      pendingSelection.current = null;
    }
  });

  const apply = (action: Action) => {
    const el = textareaRef.current;
    if (!el) return;

    const start = el.selectionStart;
    const end = el.selectionEnd;
    const selected = value.slice(start, end);

    if (action.kind === "wrap") {
      const text = selected || action.sample;
      onChange(
        value.slice(0, start) +
          action.before +
          text +
          action.after +
          value.slice(end),
      );
      const from = start + action.before.length;
      pendingSelection.current = [from, from + text.length];
      return;
    }

    if (action.kind === "link") {
      const text = selected || "link szövege";
      const url = "https://";
      onChange(value.slice(0, start) + `[${text}](${url})` + value.slice(end));
      const from = start + text.length + 3;
      pendingSelection.current = [from, from + url.length];
      return;
    }

    const lineStart = value.lastIndexOf("\n", start - 1) + 1;
    const nextNewline = value.indexOf("\n", end);
    const lineEnd = nextNewline === -1 ? value.length : nextNewline;

    const block = value.slice(lineStart, lineEnd) || action.sample;
    const prefixed = block
      .split("\n")
      .map((line, i) => {
        const prefix =
          typeof action.prefix === "function" ? action.prefix(i) : action.prefix;
        return prefix + line;
      })
      .join("\n");

    onChange(value.slice(0, lineStart) + prefixed + value.slice(lineEnd));
    pendingSelection.current = [lineStart, lineStart + prefixed.length];
  };

  const tabClass = (active: boolean) =>
    `rounded px-2.5 py-1 text-xs font-medium transition-colors duration-200 ease-smooth ${
      active
        ? "bg-surface text-foreground shadow-card"
        : "text-muted-foreground hover:text-foreground"
    }`;

  return (
    <div className="overflow-hidden rounded-control border border-line transition-[border-color,box-shadow] duration-200 ease-smooth focus-within:border-foreground focus-within:ring-2 focus-within:ring-foreground/10">
      <div className="flex flex-wrap items-center gap-1 border-b border-line bg-surface-muted px-2 py-1.5">
        {TOOLBAR.map((item) => (
          <button
            key={item.label}
            type="button"
            title={item.title}
            aria-label={item.title}
            disabled={tab === "preview"}
            onClick={() => apply(item.action)}
            className="inline-flex h-7 min-w-7 items-center justify-center rounded px-1.5 text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-line hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
          >
            {item.icon}
          </button>
        ))}

        <div className="ml-auto flex items-center gap-1 rounded bg-line/60 p-0.5">
          <button
            type="button"
            onClick={() => setTab("write")}
            aria-pressed={tab === "write"}
            className={tabClass(tab === "write")}
          >
            Írás
          </button>
          <button
            type="button"
            onClick={() => setTab("preview")}
            aria-pressed={tab === "preview"}
            className={tabClass(tab === "preview")}
          >
            Előnézet
          </button>
        </div>
      </div>

      {tab === "write" ? (
        <textarea
          id={id}
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={rows}
          placeholder={placeholder}
          className="block w-full resize-y bg-surface px-3.5 py-2.5 font-mono text-sm text-foreground placeholder:text-subtle-foreground focus:outline-none"
        />
      ) : (
        <div
          className="min-h-48 bg-surface px-3.5 py-2.5 text-[15px] leading-relaxed text-foreground"
        >
          {value.trim() ? (
            <PostContent content={value} />
          ) : (
            <p className="text-sm text-muted-foreground">
              Még nincs mit megmutatni. Válts az Írás fülre és kezdj el gépelni.
            </p>
          )}
        </div>
      )}
    </div>
  );
}
