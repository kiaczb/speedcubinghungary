/**
 * Az API elérése — egy helyen.
 *
 * ⚠️ KÉT base URL van, és ez nem elírás: `serverApiUrl()` a Docker service
 * nevére mutat (`http://api:8080`), amit csak a konténer-hálózaton belül lehet
 * feloldani; `clientApiUrl()` a böngészőből induló kérésekhez való. A
 * `NEXT_PUBLIC_*` változók BUILD IDŐBEN égnek a bundle-be, ezért módosításuk
 * után web rebuild kell.
 *
 * A két `*Fetch`/`*Json` függvény azért van, hogy a hívási helyeken ne
 * kelljen újra és újra eltalálni ugyanazt a négy dolgot: melyik base URL,
 * megy-e a JWT süti, mi lesz a nem-ok válaszból, és hogyan olvassuk ki a
 * szerver hibaüzenetét.
 */

import { unstable_rethrow } from "next/navigation";

const LOCAL_API_URL = "http://localhost:8080";

export function serverApiUrl(): string {
  return (
    process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL
  );
}

export function clientApiUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL;
}

/** A hibás válaszból dobott hiba. A `status` alapján a hívó finomíthat. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function errorFrom(response: Response): Promise<ApiError> {
  // A munkamenet egy óráig él, ami egy verseny alatt simán lejár. A nyers
  // "unauthorized" a felhasználónak semmit nem mond arról, mit kell tennie.
  if (response.status === 401 || response.status === 403) {
    return new ApiError(
      response.status,
      "Lejárt a bejelentkezésed — jelentkezz be újra.",
    );
  }

  if (response.status === 429) {
    return new ApiError(
      response.status,
      "Túl sok kérés — várj egy kicsit, és próbáld újra.",
    );
  }

  const message = await response.text().catch(() => "");
  return new ApiError(
    response.status,
    message.trim() || `A kérés sikertelen (${response.status})`,
  );
}

type ClientInit = Omit<RequestInit, "headers"> & {
  headers?: Record<string, string>;
  /**
   * JSON törzs, a Content-Type-pal együtt. FormData-hoz NE ezt használd, hanem
   * a `body`-t: a boundary miatt a Content-Type fejlécet a böngésző állítja be.
   */
  json?: unknown;
};

/**
 * Böngészőből induló API-kérés. A JWT sütit MINDIG viszi: az API
 * cookie-alapú, és ennek elfelejtése némán kijelentkezett kérést küld.
 * Nem-ok válasznál `ApiError`-t dob.
 */
export async function clientFetch(
  path: string,
  init: ClientInit = {},
): Promise<Response> {
  const { json, headers, ...rest } = init;

  const requestInit: RequestInit = { credentials: "include", ...rest };
  if (json !== undefined) {
    requestInit.body = JSON.stringify(json);
    requestInit.headers = { "Content-Type": "application/json", ...headers };
  } else if (headers) {
    requestInit.headers = headers;
  }

  const response = await fetch(`${clientApiUrl()}${path}`, requestInit);

  if (!response.ok) throw await errorFrom(response);

  return response;
}

/** Ugyanaz, mint a `clientFetch`, de a JSON választ adja vissza. */
export async function clientJson<T>(
  path: string,
  init: ClientInit = {},
): Promise<T> {
  return (await clientFetch(path, init)).json() as Promise<T>;
}

/**
 * Server Componentből induló GET.
 *
 * Hiba esetén a `fallback`-et adja vissza és naplóz. Ez szándékos: ezek az
 * oldalak backend nélkül is RENDERELŐDNEK, csak üresen — egy elérhetetlen API
 * ne cserélje le a teljes oldalt hibaképernyőre.
 */
export async function serverJson<T>(path: string, fallback: T): Promise<T> {
  try {
    const response = await fetch(`${serverApiUrl()}${path}`, {
      cache: "no-store",
    });
    if (!response.ok) return fallback;

    return (await response.json()) as T;
  } catch (error) {
    // A no-store fetch a Next belső „dinamikus renderelés” jelzését is
    // hibaként dobja — azt tovább kell engedni, nem backend-hiba.
    unstable_rethrow(error);
    console.error(`Nem sikerült elérni a backendet (${path}):`, error);
    return fallback;
  }
}
