import type { LearnNode } from "../types";

export const LEARN_HREF = "/learn";
export const ALGORITHMS_HREF = "/learn/algorithms";

export const ALGORITHMS_DESCRIPTION =
  "Válaszd ki a kirakót, a módszert és az esetet — minden esethez egy vagy több bevált algoritmust találsz.";

export function levelHref(path: string[]): string {
  return path.length === 0 ? ALGORITHMS_HREF : `${ALGORITHMS_HREF}/${path.join("/")}`;
}

export function levelApiPath(path: string[]): string {
  return `/learn/level?path=${encodeURIComponent(path.join("/"))}`;
}

export function contentSummary(node: LearnNode): string {
  if (node.childCount > 0) return `${node.childCount} elem`;
  if (node.algorithmCount > 0) return `${node.algorithmCount} algoritmus`;
  return "Még üres";
}

export function deleteNodeDescription(node: LearnNode): string {
  const contents = [
    node.childCount > 0 && `${node.childCount} alatta lévő elem, a teljes tartalmukkal`,
    node.algorithmCount > 0 && `${node.algorithmCount} algoritmus`,
  ].filter(Boolean);

  const cascade =
    contents.length > 0 ? ` Vele együtt törlődik: ${contents.join(", ")}.` : "";

  return `Biztosan törlöd a(z) „${node.title}” elemet?${cascade} Ez nem vonható vissza.`;
}

export function moveItem(ids: string[], index: number, delta: -1 | 1): string[] {
  const target = index + delta;
  if (target < 0 || target >= ids.length) return ids;

  const next = [...ids];
  [next[index], next[target]] = [next[target], next[index]];
  return next;
}
