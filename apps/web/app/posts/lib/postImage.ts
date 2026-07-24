/**
 * A poszt borítóképének teljes URL-je.
 *
 * A backend API-relatív útvonalat ad vissza (pl. "/uploads/<uuid>.webp"), amit a
 * NYILVÁNOS API base URL-jével egészítünk ki — a képet a böngésző tölti be, ezért
 * itt sosem a szerver-oldali SERVER_API_URL kell (az a Docker-belső api:8080).
 *
 * Egydomainű üzemben (Caddy) ez pl. https://test.example.com/api/uploads/x.webp,
 * fejlesztésben http://localhost:8080/uploads/x.webp.
 */
export function postImageUrl(imagePath?: string): string | null {
  if (!imagePath) return null;

  const base = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  return `${base}${imagePath}`;
}
