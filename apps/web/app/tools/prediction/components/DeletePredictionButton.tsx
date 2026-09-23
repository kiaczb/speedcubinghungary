"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { FaRegTrashAlt } from "react-icons/fa";
import { clientApiUrl } from "@/app/lib/api";

type Props = {
  predictionId: string;
  competitionName: string;
  className?: string;
};

export default function DeletePredictionButton({
  predictionId,
  competitionName,
  className = "",
}: Props) {
  const router = useRouter();
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    const confirmed = window.confirm(
      `Biztosan törlöd a(z) "${competitionName}" tippversenyt? Ez nem vonható vissza.`,
    );
    if (!confirmed) return;

    setDeleting(true);
    try {
      await fetch(`${clientApiUrl()}/predictions/${predictionId}`, {
        method: "DELETE",
        credentials: "include",
      });
      router.refresh();
    } finally {
      setDeleting(false);
    }
  };

  return (
    <button
      type="button"
      onClick={handleDelete}
      disabled={deleting}
      aria-label={`${competitionName} tippverseny törlése`}
      className={`relative z-10 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-brand-red/10 hover:text-brand-red disabled:pointer-events-none disabled:opacity-50 ${className}`}
    >
      <FaRegTrashAlt className="h-3.5 w-3.5" />
    </button>
  );
}
