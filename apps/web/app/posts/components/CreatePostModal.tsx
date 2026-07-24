"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { badgeMap, TagTypes } from "../types";

// A backend korlátaival egyezzenek, hogy a hibát ne csak a szerver jelezze:
// lásd apps/api/internal/post/upload.go és validator.go
const MAX_IMAGE_BYTES = 5 * 1024 * 1024; // 5 MB
const MIN_CONTENT_LENGTH = 20;
const ACCEPTED_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];

type Props = {
  onClose: () => void;
};

export default function CreatePostModal({ onClose }: Props) {
  const router = useRouter();
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const [title, setTitle] = useState("");
  const [tag, setTag] = useState<TagTypes>("news");
  const [content, setContent] = useState("");
  const [image, setImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);
  // Az aktuális preview objectURL-je. Ref-ben tartjuk, hogy csere/unmount esetén
  // fel tudjuk szabadítani anélkül, hogy effectben hívnánk setState-et.
  const objectUrlRef = useRef<string | null>(null);

  // Escape-re zárjon
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  // Unmountkor szabadítsuk fel az utolsó objectURL-t (különben memóriát szivárogtat).
  useEffect(() => {
    return () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
    };
  }, []);

  // A képet és az előnézetét együtt állítjuk, a korábbi objectURL-t eldobva.
  const applyImage = (file: File | null) => {
    if (objectUrlRef.current) {
      URL.revokeObjectURL(objectUrlRef.current);
      objectUrlRef.current = null;
    }
    if (file) objectUrlRef.current = URL.createObjectURL(file);

    setPreviewUrl(objectUrlRef.current);
    setImage(file);
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] ?? null;
    setError(null);

    if (!file) {
      applyImage(null);
      return;
    }
    if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
      setError("Nem támogatott képformátum. Engedélyezett: JPEG, PNG, WebP, GIF.");
      e.target.value = "";
      applyImage(null);
      return;
    }
    if (file.size > MAX_IMAGE_BYTES) {
      setError("A kép túl nagy — legfeljebb 5 MB lehet.");
      e.target.value = "";
      applyImage(null);
      return;
    }
    applyImage(file);
  };

  const clearImage = () => {
    applyImage(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!title.trim()) {
      setError("A cím nem lehet üres.");
      return;
    }
    if (content.trim().length < MIN_CONTENT_LENGTH) {
      setError(`A leírás legalább ${MIN_CONTENT_LENGTH} karakter legyen.`);
      return;
    }

    // multipart/form-data — a Content-Type fejlécet a böngésző állítja be
    // (a boundary miatt kézzel NEM szabad megadni).
    const formData = new FormData();
    formData.append("title", title.trim());
    formData.append("tag", tag);
    formData.append("content", content.trim());
    if (image) formData.append("image", image);

    setSubmitting(true);
    try {
      const res = await fetch(`${API_URL}/posts`, {
        method: "POST",
        body: formData,
        credentials: "include", // JWT cookie küldése
      });

      if (!res.ok) {
        const message = await res.text();
        throw new Error(message || "Hiba a poszt létrehozásakor");
      }

      router.refresh();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      onClick={onClose}
      role="presentation"
    >
      <div
        className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white p-6 shadow-xl dark:bg-gray-900"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-post-title"
      >
        <div className="mb-5 flex items-start justify-between">
          <h2 id="create-post-title" className="text-2xl font-bold">
            Új poszt létrehozása
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Bezárás"
            className="text-2xl leading-none text-gray-500 hover:text-gray-800 dark:hover:text-gray-200"
          >
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="post-title" className="mb-1 block text-sm font-medium">
              Cím
            </label>
            <input
              id="post-title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="A poszt címe"
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800"
            />
          </div>

          <div>
            <label htmlFor="post-tag" className="mb-1 block text-sm font-medium">
              Kategória
            </label>
            <select
              id="post-tag"
              value={tag}
              onChange={(e) => setTag(e.target.value as TagTypes)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800"
            >
              {Object.entries(badgeMap).map(([value, { label }]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="post-content" className="mb-1 block text-sm font-medium">
              Leírás
            </label>
            <textarea
              id="post-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={7}
              placeholder={"A poszt tartalma...\n\n## Alcím\n**félkövér**, *dőlt*, [link](https://...)\n\n- felsorolás"}
              className="w-full rounded-md border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800"
            />
            <p className="mt-1 text-xs text-gray-500">
              Markdown támogatott (címek, listák, táblázatok, **félkövér**…).{" "}
              {content.trim().length} / {MIN_CONTENT_LENGTH} karakter minimum
            </p>
          </div>

          <div>
            <label htmlFor="post-image" className="mb-1 block text-sm font-medium">
              Borítókép <span className="text-gray-500">(opcionális)</span>
            </label>
            <input
              id="post-image"
              ref={fileInputRef}
              type="file"
              accept={ACCEPTED_IMAGE_TYPES.join(",")}
              onChange={handleImageChange}
              className="w-full text-sm file:mr-3 file:rounded-md file:border-0 file:bg-blue-500 file:px-4 file:py-2 file:text-white hover:file:bg-blue-600"
            />
            {previewUrl && (
              <div className="mt-3">
                {/* next/image helyett sima img: a kép a saját API-nkról jön, aminek
                    a domainje környezetenként más — így nem kell remotePatterns. */}
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={previewUrl}
                  alt="Előnézet"
                  className="max-h-56 w-auto rounded-md border border-gray-200 dark:border-gray-700"
                />
                <button
                  type="button"
                  onClick={clearImage}
                  className="mt-2 text-sm text-red-600 hover:underline"
                >
                  Kép eltávolítása
                </button>
              </div>
            )}
          </div>

          {error && (
            <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">
              {error}
            </p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="rounded-md border border-gray-300 px-4 py-2 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:hover:bg-gray-800"
            >
              Mégse
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 disabled:opacity-50"
            >
              {submitting ? "Létrehozás..." : "Létrehozás"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
