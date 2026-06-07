"use client";
import { useRouter } from "next/navigation";
import React from "react";

const newPost = {
  tag: "announcement",
  title: "Új poszt címe",
  content:
    "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Nisi totam hic sit deleniti! Lorem ipsum dolor sit amet consectetur, adipisicing elit. Nostrum distinctio atque ut non neque. Iure magni veritatis dolores inventore obcaecati ut excepturi fuga magnam ad. Nisi totam hic sit deleniti!",
};

export default function CreatePostButton() {
  const router = useRouter();
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  const handleClick = async () => {
    try {
      const response = await fetch(`${API_URL}/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newPost),
        credentials: "include", // JWT cookie küldése
      });
      router.refresh();

      if (!response.ok) throw new Error("Hiba a poszt létrehozásakor");

      alert("Poszt létrehozva!");
    } catch (err) {
      console.error(err);
    }
  };
  return (
    <div className="p-3 bg-blue-500 rounded-md text-white text-center hover:cursor-pointer">
      <button onClick={handleClick}>Új poszt létrehozása</button>
    </div>
  );
}
