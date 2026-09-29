import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV !== "production";

/**
 * Az API origin-je. Test/prod alatt ez UGYANAZ a domain (a Caddy a `/api`
 * előtagot route-olja), tehát a `'self'` fedi — fejlesztésben viszont külön
 * origin (`http://localhost:8080`), amit fel kell venni MINDEN direktívába,
 * amin keresztül az API-hoz nyúlunk: a fetch-ekhez (`connect-src`) és a
 * poszt-képekhez (`img-src`, a `/uploads/*` az API-ról jön, és a `https:`
 * a http-s fejlesztői API-t nem fedi).
 */
function apiOrigin(): string | null {
  const raw = process.env.NEXT_PUBLIC_API_URL;
  if (!raw) return null;

  try {
    return new URL(raw).origin;
  } catch {
    return null;
  }
}

const selfAndApi = ["'self'", apiOrigin()].filter(Boolean).join(" ");

/**
 * ⚠️ `script-src` `'unsafe-inline'`-t tartalmaz, és ez SZÁNDÉKOS.
 *
 * A szigorú, nonce-alapú változat a Next 16-ban egy `proxy.ts`-t és
 * MINDEN oldal dinamikus renderelését igényli (a nonce kérésenként születik,
 * statikusan generált oldalba nem lehet beleírni). Ez a site-on statikus
 * renderelést, CDN-gyorsítótárazhatóságot és a `/tools/special`-féle oldalak
 * jelenlegi működését cserélné le — külön döntés, nem ennek a fejlécnek a
 * mellékhatása.
 *
 * Amit így is megkapunk, és ami itt tényleg számít: külső eredetű script,
 * `connect-src` és `frame-ancestors` tiltása. A poszt-tartalom amúgy is
 * escape-elve renderelődik (nincs `rehype-raw`, lásd CLAUDE.md §6), tehát ez
 * védelmi réteg, nem az egyetlen védelem.
 *
 * `img-src https:`: a WCA profilképek hosztja nincs a kódban rögzítve és
 * változhat; a kép nem futtatható tartalom, ezért itt a megengedőbb szabály
 * ára kicsi, a haszna (nem tűnik el némán minden avatar) nagy.
 */
const contentSecurityPolicy = [
  "default-src 'self'",
  "base-uri 'self'",
  "object-src 'none'",
  "frame-ancestors 'none'",
  "form-action 'self'",
  `script-src 'self' 'unsafe-inline'${isDev ? " 'unsafe-eval'" : ""}`,
  "style-src 'self' 'unsafe-inline'",
  `img-src ${selfAndApi} data: blob: https:`,
  "font-src 'self' data:",
  `connect-src ${selfAndApi}`,
  // Fejlesztésben nincs HTTPS, és ez a localhost:8080-as API-hívásokat törné el.
  ...(isDev ? [] : ["upgrade-insecure-requests"]),
].join("; ");

const securityHeaders = [
  { key: "Content-Security-Policy", value: contentSecurityPolicy },
  // A frame-ancestors modern böngészőben elég, ez a régiek miatt marad.
  { key: "X-Frame-Options", value: "DENY" },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  {
    key: "Permissions-Policy",
    value: "camera=(), microphone=(), geolocation=(), payment=()",
  },
];

const nextConfig: NextConfig = {
  // Emit a self-contained production server at .next/standalone so the prod
  // Docker image can run without node_modules. No effect on `next dev`.
  output: "standalone",
  // Turbopack (default in Next 16) honors this for file watching. Polling is
  // needed because native FS events don't propagate over Docker bind mounts
  // in the dev container. Replaces the old webpack `watchOptions.poll`.
  watchOptions: {
    pollIntervalMs: 1000,
  },
  async headers() {
    return [{ source: "/:path*", headers: securityHeaders }];
  },
};

export default nextConfig;
