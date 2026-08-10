/**
 * Az API bázis URL-jei.
 *
 * Két külön URL van, és NEM cserélhetők fel:
 *
 * - `serverApiUrl()` — szerver oldali fetch-hez (Server Component, route handler).
 *   A Docker service nevére mutat (`http://api:8080`), amit csak a konténer-
 *   hálózaton belül lehet feloldani.
 * - `clientApiUrl()` — böngészőből induló fetch-hez. A `NEXT_PUBLIC_*` változók
 *   build időben égnek bele a bundle-be, ezért módosításuk után web rebuild kell.
 *
 * Korábban ez a két kifejezés hét helyen volt bemásolva; a fallback-sorrend itt
 * változatlan maradt.
 */

const LOCAL_API_URL = "http://localhost:8080";

/** Szerver oldali fetch-ekhez. Böngészőből ne hívd — az `api` host nem oldható fel. */
export function serverApiUrl(): string {
  return (
    process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL
  );
}

/** Böngészőből induló fetch-ekhez (`"use client"` komponensek). */
export function clientApiUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL;
}
