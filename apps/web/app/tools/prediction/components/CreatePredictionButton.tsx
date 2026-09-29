"use client";
import { useState } from "react";
import dynamic from "next/dynamic";
import { FaPlus } from "react-icons/fa";
import Button from "@/app/components/ui/Button";

// Csak adminnak kell — külön chunkban, hogy a lista bundle-je ne vigye
// magával minden látogatónak (lásd posts/components/CreatePostButton.tsx).
const CreatePredictionModal = dynamic(
  () => import("./CreatePredictionModal"),
  { ssr: false },
);

export default function CreatePredictionButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        size="sm"
        onClick={() => setOpen(true)}
        icon={<FaPlus className="h-3 w-3" />}
      >
        Új tippverseny
      </Button>

      {open && <CreatePredictionModal onClose={() => setOpen(false)} />}
    </>
  );
}
