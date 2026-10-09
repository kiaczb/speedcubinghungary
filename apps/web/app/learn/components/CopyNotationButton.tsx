"use client";
import { useEffect, useState } from "react";
import { FaCheck, FaRegCopy } from "react-icons/fa";

type Props = {
  notation: string;
};

export default function CopyNotationButton({ notation }: Props) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return;
    const timer = window.setTimeout(() => setCopied(false), 1500);
    return () => window.clearTimeout(timer);
  }, [copied]);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(notation);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  };

  const label = copied ? "Kimásolva" : "Algoritmus másolása";

  return (
    <button
      type="button"
      onClick={copy}
      aria-label={label}
      title={label}
      className={`inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-control transition-colors duration-200 ease-smooth hover:bg-surface-muted ${
        copied ? "text-brand-green" : "text-subtle-foreground hover:text-foreground"
      }`}
    >
      {copied ? <FaCheck className="h-3.5 w-3.5" /> : <FaRegCopy className="h-3.5 w-3.5" />}
    </button>
  );
}
