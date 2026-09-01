/**
 * A poszt tartalmának rövid, sima szöveges előnézete a listákhoz/kártyákhoz.
 *
 * A content markdown, de a listában nem rendereljük — csak egy pár szavas
 * ízelítő kell, markdown-jelek NÉLKÜL (különben a `#`, `*`, `|` stb. ott
 * éktelenkedne). Ez nem teljes markdown-parser, csak a gyakori jelöléseket
 * távolítja el és normalizálja a whitespace-t.
 */
export function postPreview(content: string, maxLength = 160): string {
  const stripped = content
    // képek ![alt](url) -> alt
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, "$1")
    // linkek [szöveg](url) -> szöveg
    .replace(/\[([^\]]*)\]\([^)]*\)/g, "$1")
    // táblázat-elválasztó sorok (| --- | --- |)
    .replace(/^\s*\|?[\s:|-]+\|?\s*$/gm, "")
    // fejléc-jelek, idézet-jelek, listajelek a sor elején
    .replace(/^\s{0,3}(#{1,6}|>|[-*+])\s+/gm, "")
    // félkövér/dőlt/kód jelek, tábla-pipe-ok
    .replace(/[*_`|]/g, "")
    // maradék whitespace normalizálása
    .replace(/\s+/g, " ")
    .trim();

  if (stripped.length <= maxLength) return stripped;
  // Szó határon vágjunk, ne a szó közepén.
  return stripped.slice(0, maxLength).replace(/\s+\S*$/, "") + "…";
}
