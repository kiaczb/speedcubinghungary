"use client";
import { useState } from "react";
import CreatePostModal from "./CreatePostModal";

export default function CreatePostButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="rounded-md bg-blue-500 p-3 text-center text-white hover:cursor-pointer hover:bg-blue-600"
      >
        Új poszt létrehozása
      </button>

      {open && <CreatePostModal onClose={() => setOpen(false)} />}
    </>
  );
}
