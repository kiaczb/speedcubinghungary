"use client";

import { useRef } from "react";
import { controlClass } from "@/app/components/ui/Field";
import type { TemplateKind, TextSegment } from "../../../lib/template/model";
import { VARIABLE_GROUPS, variablesFor } from "../../../lib/template/variables";
import type { VariableKey } from "../../../lib/template/variables";
import TextContentEditor from "./TextContentEditor";
import type { TextContentEditorHandle } from "./TextContentEditor";

export function VariableSelect({ kind, onPick }: { kind: TemplateKind; onPick: (key: VariableKey) => void }) {
  const variables = variablesFor(kind);
  return (
    <select
      aria-label="Adatmező beszúrása"
      value=""
      onChange={(event) => {
        if (event.target.value) onPick(event.target.value as VariableKey);
      }}
      className={`${controlClass} px-2.5 py-1.5 text-muted-foreground`}
    >
      <option value="">Adatmező beszúrása a kurzorhoz…</option>
      {VARIABLE_GROUPS.map((group) => {
        const inGroup = variables.filter((v) => v.group === group);
        if (inGroup.length === 0) return null;
        return (
          <optgroup key={group} label={group}>
            {inGroup.map((v) => (
              <option key={v.key} value={v.key}>
                {v.label}
              </option>
            ))}
          </optgroup>
        );
      })}
    </select>
  );
}

export default function SegmentsField({
  id,
  kind,
  segments,
  onChange,
}: {
  id: string;
  kind: TemplateKind;
  segments: TextSegment[];
  onChange: (segments: TextSegment[]) => void;
}) {
  const editor = useRef<TextContentEditorHandle>(null);
  return (
    <div className="flex flex-col gap-2">
      <TextContentEditor key={id} id={id} ref={editor} segments={segments} onChange={onChange} />
      <VariableSelect kind={kind} onPick={(key) => editor.current?.insertVariable(key)} />
    </div>
  );
}
