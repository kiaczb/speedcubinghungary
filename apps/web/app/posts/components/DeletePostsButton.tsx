"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { FaRegTrashAlt } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import { clientApiUrl } from "@/app/lib/api";

type Props = {
  selectedIds: string[];
  onDeleted?: () => void;
};

export default function DeletePostsButton({ selectedIds, onDeleted }: Props) {
  const router = useRouter();
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    if (selectedIds.length === 0) return;

    const confirmed = window.confirm(
      selectedIds.length === 1
        ? "Biztosan törlöd a kijelölt posztot? Ez nem vonható vissza."
        : `Biztosan törlöd a kijelölt ${selectedIds.length} posztot? Ez nem vonható vissza.`,
    );
    if (!confirmed) return;

    setDeleting(true);
    try {
      await fetch(`${clientApiUrl()}/posts`, {
        method: "DELETE",
        body: JSON.stringify({ ids: selectedIds }),
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      });

      router.refresh();
      onDeleted?.();
    } finally {
      setDeleting(false);
    }
  };

  return (
    <Button
      variant="danger"
      size="sm"
      disabled={selectedIds.length === 0 || deleting}
      onClick={handleDelete}
      icon={<FaRegTrashAlt className="h-3 w-3" />}
    >
      {deleting ? "Törlés…" : `Törlés (${selectedIds.length})`}
    </Button>
  );
}
