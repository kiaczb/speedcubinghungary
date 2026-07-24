export type PostsItem = {
  id: string;
  tag: TagTypes;
  title: string;
  content: string;
  /** A borítókép API-relatív útvonala (pl. "/uploads/<uuid>.webp"); az API ""-t ad kép nélküli poszthoz. */
  image_path?: string;
  created_at: string;
};

export const badgeMap = {
  news: { label: "Hír", color: "bg-blue-500" },
  event: { label: "Esemény", color: "bg-green-500" },
  competition: { label: "Verseny", color: "bg-red-500" },
  announcement: { label: "Közlemény", color: "bg-yellow-500" },
  // A weekly-report job által generált heti riport (post_tag 'report').
  // Kötelező bejegyzés: a badge-et a tag alapján keressük ki, enélkül a
  // riport posztok renderelése elszállna.
  report: { label: "Heti riport", color: "bg-purple-500" },
} as const;

export type TagTypes = keyof typeof badgeMap;
