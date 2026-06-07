"use client";
import { useRouter } from "next/navigation";

type Props = {
  selectedIds: string[];
  onDeleted?: () => void;
};

export default function DeletePostsButton({ selectedIds, onDeleted }: Props) {
  const router = useRouter();
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const handleDelete = async () => {
    if (selectedIds.length === 0) return;

    await fetch(`${API_URL}/posts`, {
      method: "DELETE",
      body: JSON.stringify({ ids: selectedIds }),
      headers: { "Content-Type": "application/json" },
      credentials: "include",
    });

    router.refresh();
    onDeleted?.();
  };

  return (
    <button
      disabled={selectedIds.length === 0}
      onClick={handleDelete}
      className="mb-4 bg-red-500 text-white px-4 py-2 rounded disabled:opacity-50"
    >
      Törlés ({selectedIds.length})
    </button>
  );
}
