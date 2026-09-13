// A videó felismerése EGY helyen történik, hogy a megjelenítés ne az URL
// alakjától függjön. Ma egy külső link (jellemzően YouTube) az egyetlen forrás;
// ha később feltöltött videó, felhőtárhely vagy több videó jön, elég ezt a
// réteget kicserélni — a komponensek a `VideoEmbed`-et fogyasztják, nem az URL-t.

export type VideoEmbed =
  | { kind: "youtube"; src: string; url: string }
  | { kind: "file"; src: string; url: string }
  | { kind: "link"; url: string };

const FILE_EXTENSIONS = [".mp4", ".webm", ".ogg", ".ogv", ".mov"];

/**
 * `null`, ha nincs videó vagy az URL nem használható. A hívó ilyenkor NE
 * rendereljen üres videókeretet.
 */
export function parseVideo(rawUrl?: string | null): VideoEmbed | null {
  const url = rawUrl?.trim();
  if (!url) return null;

  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    return null;
  }

  // Csak http(s): az érték iframe/anchor forrásába kerül, így a `javascript:`
  // és `data:` sémák nem juthatnak el a böngészőig. A backend ugyanezt
  // ellenőrzi mentéskor (learn/validator.go) — ez a második védvonal.
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return null;

  const youTubeId = youTubeVideoId(parsed);
  if (youTubeId) {
    return {
      kind: "youtube",
      // A nocookie domain nem tesz le hirdetési sütit, amíg a néző nem indítja el.
      src: `https://www.youtube-nocookie.com/embed/${youTubeId}`,
      url,
    };
  }

  const path = parsed.pathname.toLowerCase();
  if (FILE_EXTENSIONS.some((extension) => path.endsWith(extension))) {
    return { kind: "file", src: url, url };
  }

  // Ismeretlen szolgáltató: nem ágyazzuk be, csak linkeljük — így egy Vimeo-
  // vagy Drive-link is használható marad, csak nem beágyazva.
  return { kind: "link", url };
}

function youTubeVideoId(parsed: URL): string | null {
  const host = parsed.hostname.replace(/^www\./, "").toLowerCase();

  if (host === "youtu.be") {
    return sanitizeId(parsed.pathname.slice(1));
  }

  if (host !== "youtube.com" && host !== "m.youtube.com" && host !== "youtube-nocookie.com") {
    return null;
  }

  if (parsed.pathname === "/watch") {
    return sanitizeId(parsed.searchParams.get("v") ?? "");
  }

  // /embed/<id>, /shorts/<id>, /live/<id> — mind ugyanarra az id-re mutat.
  const match = parsed.pathname.match(/^\/(embed|shorts|live)\/([^/]+)/);
  if (match) return sanitizeId(match[2]);

  return null;
}

// A YouTube videóazonosítója betű/szám/kötőjel/aláhúzás — bármi más gyanús,
// és nem kerülhet bele az iframe URL-jébe.
function sanitizeId(id: string): string | null {
  const trimmed = id.trim();
  return /^[A-Za-z0-9_-]{5,20}$/.test(trimmed) ? trimmed : null;
}
