export type PostsItem = {
  id: string;
  tag: TagTypes;
  title: string;
  content: string;
  created_at: string;
};

export const badgeMap = {
  news: { label: "Hír", color: "bg-blue-500" },
  event: { label: "Esemény", color: "bg-green-500" },
  competition: { label: "Verseny", color: "bg-red-500" },
  announcement: { label: "Közlemény", color: "bg-yellow-500" },
} as const;

export type TagTypes = keyof typeof badgeMap;
