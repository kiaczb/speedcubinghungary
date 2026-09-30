import { clientApiUrl } from "@/app/lib/api";

export const MAX_POST_IMAGES = 10;

export function postImageUrl(imagePath?: string): string | null {
  if (!imagePath) return null;

  return `${clientApiUrl()}${imagePath}`;
}

export function postImageUrls(imagePaths?: string[]): string[] {
  if (!imagePaths) return [];

  const base = clientApiUrl();
  return imagePaths.filter(Boolean).map((path) => `${base}${path}`);
}

export function postCoverUrl(imagePaths?: string[]): string | null {
  return postImageUrl(imagePaths?.[0]);
}
