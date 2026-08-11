"use client";
import { useState } from "react";
import dynamic from "next/dynamic";
import { FaPlus } from "react-icons/fa";
import Button from "@/app/components/ui/Button";

// A szerkesztő csak adminnak kell — külön chunkban, hogy a /posts oldal
// bundle-je ne vigye magával minden látogatónak.
const CreatePostModal = dynamic(() => import("./CreatePostModal"), {
  ssr: false,
});

export default function CreatePostButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        size="sm"
        onClick={() => setOpen(true)}
        icon={<FaPlus className="h-3 w-3" />}
      >
        Új poszt
      </Button>

      {open && <CreatePostModal onClose={() => setOpen(false)} />}
    </>
  );
}
