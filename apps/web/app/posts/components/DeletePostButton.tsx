"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { FaRegTrashAlt } from "react-icons/fa";
import ConfirmModal from "@/app/components/ui/ConfirmModal";
import { clientFetch } from "@/app/lib/api";

type Props = {
  postId: string;
  title: string;
};

export default function DeletePostButton({ postId, title }: Props) {
  const router = useRouter();
  const [confirming, setConfirming] = useState(false);

  const handleConfirm = async () => {
    await clientFetch("/posts", { method: "DELETE", json: { ids: [postId] } });
    router.refresh();
  };

  return (
    <>
      <button
        type="button"
        onClick={() => setConfirming(true)}
        aria-label={`${title} törlése`}
        className="relative z-10 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-brand-red/10 hover:text-brand-red"
      >
        <FaRegTrashAlt className="h-4 w-4" />
      </button>

      {confirming && (
        <ConfirmModal
          title="Poszt törlése"
          description={`Biztosan törlöd a(z) "${title}" posztot? Ez nem vonható vissza.`}
          onConfirm={handleConfirm}
          onClose={() => setConfirming(false)}
        />
      )}
    </>
  );
}
