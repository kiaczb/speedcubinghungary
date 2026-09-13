import { clientApiUrl } from "@/app/lib/api";

/**
 * Egyezzen a backend korlátjával: apps/api/internal/post/upload.go (MaxImages).
 */
export const MAX_POST_IMAGES = 10;

/**
 * A backend API-relatív útvonalat ad vissza (pl. "/uploads/<uuid>.webp"), amit a
 * NYILVÁNOS API base URL-jével egészítünk ki — a képet a böngésző tölti be, ezért
 * itt sosem a szerver-oldali SERVER_API_URL kell (az a Docker-belső api:8080).
 */
export function postImageUrl(imagePath?: string): string | null {
  if (!imagePath) return null;

  return `${clientApiUrl()}${imagePath}`;
}

/**
 * A poszt galériájának teljes URL-listája, a tárolt SORRENDET megtartva — ez a
 * carousel bemenete.
 */
export function postImageUrls(imagePaths?: string[]): string[] {
  if (!imagePaths) return [];

  const base = clientApiUrl();
  return imagePaths.filter(Boolean).map((path) => `${base}${path}`);
}

/**
 * A borítókép (a galéria első képe), vagy null ha a poszthoz nincs kép. A
 * listák és a NewsCard ezt mutatják, nem az egész galériát.
 */
export function postCoverUrl(imagePaths?: string[]): string | null {
  return postImageUrl(imagePaths?.[0]);
}
