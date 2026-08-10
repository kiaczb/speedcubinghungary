"use client";
import { useState } from "react";
import { FaPlus } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import CreatePostModal from "./CreatePostModal";

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
