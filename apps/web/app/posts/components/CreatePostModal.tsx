"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { FaArrowLeft, FaArrowRight } from "react-icons/fa";
import { badgeMap, PostsItem, TagTypes } from "../types";
import { MAX_POST_IMAGES, postImageUrl } from "../lib/postImage";
import { clientApiUrl } from "@/app/lib/api";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";
import Modal from "@/app/components/ui/Modal";
import MarkdownEditor from "./MarkdownEditor";

// Egyezzenek a backend korlátaival: apps/api/internal/post/upload.go + validator.go
const MAX_IMAGE_BYTES = 5 * 1024 * 1024; // 5 MB
const MIN_CONTENT_LENGTH = 20;
const ACCEPTED_IMAGE_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
];


type GalleryItem =
  | { kind: "existing"; path: string; url: string }
  | { kind: "new"; file: File; url: string };

type Props = {
  onClose: () => void;
  // Ha meg van adva, a modal SZERKESZTÉS módban indul (PUT /posts/{id}),
  // különben új posztot hoz létre (POST /posts).
  post?: PostsItem;
};

export default function CreatePostModal({ onClose, post }: Props) {
  const router = useRouter();
  const API_URL = clientApiUrl();

  const isEdit = Boolean(post);

  const [title, setTitle] = useState(post?.title ?? "");
  const [tag, setTag] = useState<TagTypes>(post?.tag ?? "news");
  const [content, setContent] = useState(post?.content ?? "");
  const [images, setImages] = useState<GalleryItem[]>(() =>
    (post?.image_paths ?? []).map((path) => ({
      kind: "existing" as const,
      path,
      url: postImageUrl(path) ?? "",
    })),
  );
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const titleInputRef = useRef<HTMLInputElement>(null);

  const objectUrlsRef = useRef<Set<string>>(new Set());

  const contentLength = content.trim().length;
  const remainingSlots = MAX_POST_IMAGES - images.length;

  useEffect(() => {
    const urls = objectUrlsRef.current;
    return () => {
      urls.forEach((url) => URL.revokeObjectURL(url));
    };
  }, []);

  const handleFilesChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? []);

    e.target.value = "";
    setError(null);
    if (files.length === 0) return;

    const accepted: GalleryItem[] = [];
    let rejectedType = false;
    let rejectedSize = false;
    let rejectedCount = false;

    for (const file of files) {
      if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
        rejectedType = true;
        continue;
      }
      if (file.size > MAX_IMAGE_BYTES) {
        rejectedSize = true;
        continue;
      }
      if (images.length + accepted.length >= MAX_POST_IMAGES) {
        rejectedCount = true;
        break;
      }

      const url = URL.createObjectURL(file);
      objectUrlsRef.current.add(url);
      accepted.push({ kind: "new", file, url });
    }

    if (accepted.length > 0) setImages([...images, ...accepted]);

    if (rejectedCount) {
      setError(`Legfeljebb ${MAX_POST_IMAGES} kép csatolható a poszthoz.`);
    } else if (rejectedSize) {
      setError("A kihagyott kép(ek) túl nagyok — legfeljebb 5 MB lehet egy kép.");
    } else if (rejectedType) {
      setError(
        "Nem támogatott képformátum. Engedélyezett: JPEG, PNG, WebP, GIF.",
      );
    }
  };

  const removeImage = (index: number) => {
    const item = images[index];
    if (item.kind === "new") {
      URL.revokeObjectURL(item.url);
      objectUrlsRef.current.delete(item.url);
    }
    setImages(images.filter((_, i) => i !== index));
    setError(null);
  };

  // A sorrend maga az adat: az első kép a borítókép, és a poszt oldalán a
  // carousel is ebben a sorrendben lapoz.
  const moveImage = (index: number, delta: number) => {
    const target = index + delta;
    if (target < 0 || target >= images.length) return;

    const reordered = [...images];
    [reordered[index], reordered[target]] = [
      reordered[target],
      reordered[index],
    ];
    setImages(reordered);
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


    for (const item of images) {
      if (item.kind === "new") formData.append("images", item.file);
    }

    if (isEdit) {

      formData.append(
        "image_order",
        JSON.stringify(
          images.map((item) => (item.kind === "existing" ? item.path : "new")),
        ),
      );
    }

    setSubmitting(true);
    try {
      const res = await fetch(
        isEdit ? `${API_URL}/posts/${post!.id}` : `${API_URL}/posts`,
        {
          method: isEdit ? "PUT" : "POST",
          body: formData,
          credentials: "include", // JWT cookie küldése
        },
      );

      if (!res.ok) {
        const message = await res.text();
        throw new Error(
          message ||
            (isEdit ? "Hiba a poszt mentésekor" : "Hiba a poszt létrehozásakor"),
        );
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
    <Modal
      title={isEdit ? "Poszt szerkesztése" : "Új poszt létrehozása"}
      onClose={onClose}
      closeDisabled={submitting}
      initialFocusRef={titleInputRef}
    >
      <form onSubmit={handleSubmit} className="space-y-5 px-6 py-5">
        <Field label="Cím" htmlFor="post-title">
          <input
            id="post-title"
            ref={titleInputRef}
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="A poszt címe"
            className={controlClass}
          />
        </Field>

        <Field label="Kategória" htmlFor="post-tag">
          <select
            id="post-tag"
            value={tag}
            onChange={(e) => setTag(e.target.value as TagTypes)}
            className={controlClass}
          >
            {Object.entries(badgeMap).map(([value, { label }]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </Field>

        <Field
          label="Leírás"
          htmlFor="post-content"
          hint={
            <span
              className={
                contentLength < MIN_CONTENT_LENGTH
                  ? "font-medium text-brand-red"
                  : "text-muted-foreground"
              }
            >
              <span className="tabular font-mono">{contentLength}</span> /{" "}
              {MIN_CONTENT_LENGTH} karakter minimum
            </span>
          }
        >
          <MarkdownEditor
            id="post-content"
            value={content}
            onChange={setContent}
            rows={12}
            placeholder={"A poszt tartalma…\n\nJelöld ki a szöveget, és használd a fenti gombokat."}
          />
        </Field>

        <Field
          label="Képek"
          htmlFor="post-images"
          optional
          hint={
            <span className="text-muted-foreground">
              <span className="tabular font-mono">{images.length}</span> /{" "}
              {MAX_POST_IMAGES} kép
            </span>
          }
        >
          <input
            id="post-images"
            ref={fileInputRef}
            type="file"
            multiple
            accept={ACCEPTED_IMAGE_TYPES.join(",")}
            onChange={handleFilesChange}
            disabled={remainingSlots <= 0}
            className="w-full cursor-pointer text-sm text-muted-foreground file:mr-3 file:cursor-pointer file:rounded-control file:border file:border-line file:bg-surface-muted file:px-4 file:py-2 file:text-sm file:font-medium file:text-foreground hover:file:bg-line disabled:cursor-not-allowed disabled:opacity-50"
          />

          {images.length > 0 && (
            <>
              <p className="mt-2 text-xs text-muted-foreground">
                Az első kép a borítókép — ez látszik a listákban. A poszt
                oldalán a képek ebben a sorrendben lapozhatók.
              </p>

              <ul className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3">
                {images.map((item, index) => (
                  <li
                    key={item.url}
                    className="relative overflow-hidden rounded-control border border-line bg-surface-muted"
                  >
                    {/* Sima img: a kép domainje környezetenként más — így nem kell remotePatterns. */}
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={item.url}
                      alt={`${index + 1}. kép előnézete`}
                      className="aspect-video w-full object-cover"
                    />

                    <span className="tabular absolute left-1.5 top-1.5 rounded-control bg-surface-inverse/75 px-1.5 py-0.5 font-mono text-[11px] text-white">
                      {index === 0 ? "Borító" : index + 1}
                    </span>

                    <button
                      type="button"
                      onClick={() => removeImage(index)}
                      aria-label={`${index + 1}. kép eltávolítása`}
                      className="absolute right-1.5 top-1.5 inline-flex h-6 w-6 items-center justify-center rounded-full bg-surface-inverse/75 text-sm leading-none text-white transition-colors duration-200 ease-smooth hover:bg-brand-red"
                    >
                      ×
                    </button>

                    <div className="flex justify-center gap-1 border-t border-line bg-surface p-1">
                      <button
                        type="button"
                        onClick={() => moveImage(index, -1)}
                        disabled={index === 0}
                        aria-label={`${index + 1}. kép előrébb`}
                        className="inline-flex h-7 w-7 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
                      >
                        <FaArrowLeft className="h-3 w-3" />
                      </button>
                      <button
                        type="button"
                        onClick={() => moveImage(index, 1)}
                        disabled={index === images.length - 1}
                        aria-label={`${index + 1}. kép hátrébb`}
                        className="inline-flex h-7 w-7 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
                      >
                        <FaArrowRight className="h-3 w-3" />
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </>
          )}
        </Field>

        {error && (
          <p
            role="alert"
            className="rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong"
          >
            {error}
          </p>
        )}

        <div className="flex justify-end gap-3 border-t border-line pt-5">
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={submitting}
          >
            Mégse
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting
              ? isEdit
                ? "Mentés…"
                : "Létrehozás…"
              : isEdit
                ? "Mentés"
                : "Létrehozás"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
