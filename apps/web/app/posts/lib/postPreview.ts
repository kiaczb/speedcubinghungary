export function postPreview(content: string, maxLength = 160): string {
  const stripped = content
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, "$1")
    .replace(/\[([^\]]*)\]\([^)]*\)/g, "$1")
    .replace(/^\s*\|?[\s:|-]+\|?\s*$/gm, "")
    .replace(/^\s{0,3}(#{1,6}|>|[-*+])\s+/gm, "")
    .replace(/[*_`|]/g, "")
    .replace(/\s+/g, " ")
    .trim();

  if (stripped.length <= maxLength) return stripped;
  return stripped.slice(0, maxLength).replace(/\s+\S*$/, "") + "…";
}
