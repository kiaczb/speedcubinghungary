"use client";
import { useEffect, useRef, useState } from "react";
import { FaCheck, FaRegCopy } from "react-icons/fa";
import Button from "@/app/components/ui/Button";

type Props = {
  notation: string;
};

export default function CopyNotationButton({ notation }: Props) {
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(notation);
      setCopied(true);

      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    } catch {
      // A vágólap böngészőtől/jogosultságtól függ (nem HTTPS, letiltott
      // engedély). Ilyenkor a jelölés marad el, a nézet nem törik el —
      // az algoritmus továbbra is kijelölhető és másolható kézzel.
    }
  };

  return (
    <Button
      type="button"
      variant="secondary"
      size="sm"
      onClick={handleCopy}
      icon={
        copied ? (
          <FaCheck className="h-3 w-3" />
        ) : (
          <FaRegCopy className="h-3 w-3" />
        )
      }
    >
      {copied ? "Másolva" : "Másolás"}
    </Button>
  );
}
