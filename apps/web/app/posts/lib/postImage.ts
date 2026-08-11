import { clientApiUrl } from "@/app/lib/api";

/**
 * A backend API-relatív útvonalat ad vissza (pl. "/uploads/<uuid>.webp"), amit a
 * NYILVÁNOS API base URL-jével egészítünk ki — a képet a böngésző tölti be, ezért
 * itt sosem a szerver-oldali SERVER_API_URL kell (az a Docker-belső api:8080).
 */
export function postImageUrl(imagePath?: string): string | null {
  if (!imagePath) return null;

  return `${clientApiUrl()}${imagePath}`;
}
