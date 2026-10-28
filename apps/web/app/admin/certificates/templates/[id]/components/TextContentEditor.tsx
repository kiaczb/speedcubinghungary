"use client";

import { useEffect, useImperativeHandle, useRef } from "react";
import type { Ref } from "react";
import type { TextSegment } from "../../../lib/template/model";
import { variableLabel } from "../../../lib/template/variables";
import type { VariableKey } from "../../../lib/template/variables";

export type TextContentEditorHandle = {
  insertVariable: (key: VariableKey) => void;
  focus: () => void;
};

const CHIP_CLASS =
  "mx-px rounded-sm bg-surface-muted px-1 font-medium text-foreground underline decoration-line-strong decoration-dotted underline-offset-4";

function render(root: HTMLElement, segments: TextSegment[]) {
  root.replaceChildren();
  for (const segment of segments) {
    if ("variable" in segment) {
      const chip = document.createElement("span");
      chip.contentEditable = "false";
      chip.dataset.variable = segment.variable;
      chip.className = CHIP_CLASS;
      chip.textContent = variableLabel(segment.variable);
      root.append(chip);
      continue;
    }
    segment.text.split("\n").forEach((part, i) => {
      if (i > 0) root.append(document.createElement("br"));
      if (part) root.append(document.createTextNode(part));
    });
  }
}

function read(root: HTMLElement): TextSegment[] {
  const segments: TextSegment[] = [];
  const pushText = (text: string) => {
    if (!text) return;
    const last = segments.at(-1);
    if (last && "text" in last) last.text += text;
    else segments.push({ text });
  };
  const walk = (node: Node, first: boolean) => {
    if (node.nodeType === Node.TEXT_NODE) {
      pushText((node.textContent ?? "").replace(/ /g, " "));
      return;
    }
    if (!(node instanceof HTMLElement)) return;
    if (node.dataset.variable) {
      segments.push({ variable: node.dataset.variable as VariableKey });
      return;
    }
    if (node.tagName === "BR") {
      pushText("\n");
      return;
    }
    if ((node.tagName === "DIV" || node.tagName === "P") && !first) pushText("\n");
    node.childNodes.forEach((child, i) => walk(child, i === 0));
  };
  root.childNodes.forEach((child, i) => walk(child, i === 0));
  const last = segments.at(-1);
  if (last && "text" in last) last.text = last.text.replace(/\n$/, "");
  return segments.filter((s) => !("text" in s) || s.text !== "");
}

export default function TextContentEditor({
  segments,
  onChange,
  id,
  ref,
}: {
  segments: TextSegment[];
  onChange: (segments: TextSegment[]) => void;
  id?: string;
  ref?: Ref<TextContentEditorHandle>;
}) {
  const rootRef = useRef<HTMLDivElement>(null);
  const emitted = useRef<string>("");

  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;
    const serialized = JSON.stringify(segments);
    if (serialized === emitted.current) return;
    emitted.current = serialized;
    render(root, segments);
  }, [segments]);

  const emit = () => {
    const root = rootRef.current;
    if (!root) return;
    const next = read(root);
    emitted.current = JSON.stringify(next);
    onChange(next);
  };

  useImperativeHandle(ref, () => ({
    focus: () => rootRef.current?.focus(),
    insertVariable: (key) => {
      const root = rootRef.current;
      if (!root) return;
      const chip = document.createElement("span");
      chip.contentEditable = "false";
      chip.dataset.variable = key;
      chip.className = CHIP_CLASS;
      chip.textContent = variableLabel(key);

      const selection = window.getSelection();
      const range = selection && selection.rangeCount > 0 ? selection.getRangeAt(0) : null;
      if (range && root.contains(range.commonAncestorContainer)) {
        range.deleteContents();
        range.insertNode(chip);
        range.setStartAfter(chip);
        range.collapse(true);
        selection?.removeAllRanges();
        selection?.addRange(range);
      } else {
        root.append(chip);
      }
      emit();
      root.focus();
    },
  }));

  return (
    <div
      id={id}
      ref={rootRef}
      contentEditable
      suppressContentEditableWarning
      role="textbox"
      aria-multiline="true"
      onInput={emit}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          event.preventDefault();
          document.execCommand("insertLineBreak");
        }
      }}
      onPaste={(event) => {
        event.preventDefault();
        document.execCommand("insertText", false, event.clipboardData.getData("text/plain"));
      }}
      className="min-h-20 w-full whitespace-pre-wrap break-words rounded-control border border-line bg-surface px-3 py-2 text-sm leading-7 text-foreground focus:border-foreground focus:outline-none focus:ring-2 focus:ring-foreground/10"
    />
  );
}
