"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { badgeMap, PostsItem, TagTypes } from "../types";
import { postImageUrl } from "../lib/postImage";
import { clientApiUrl } from "@/app/lib/api";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";

// A backend korlátaival egyezzenek, hogy a hibát ne csak a szerver jelezze:
// lásd apps/api/internal/post/upload.go és validator.go
const MAX_IMAGE_BYTES = 5 * 1024 * 1024; // 5 MB
const MIN_CONTENT_LENGTH = 20;
const ACCEPTED_IMAGE_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
];

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
  const existingImageUrl = postImageUrl(post?.image_path);

  const [title, setTitle] = useState(post?.title ?? "");
  const [tag, setTag] = useState<TagTypes>(post?.tag ?? "news");
  const [content, setContent] = useState(post?.content ?? "");
  const [image, setImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  // Szerkesztéskor: igaz, ha a felhasználó a MEGLÉVŐ képet törölni akarja
  // (új kép feltöltése nélkül). Ilyenkor remove_image=true megy a backendnek.
  const [removeExisting, setRemoveExisting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const titleInputRef = useRef<HTMLInputElement>(null);
  // Az aktuális preview objectURL-je. Ref-ben tartjuk, hogy csere/unmount esetén
  // fel tudjuk szabadítani anélkül, hogy effectben hívnánk setState-et.
  const objectUrlRef = useRef<string | null>(null);

  // Az előnézetben megjelenő kép: az újonnan választott fájl elsőbbséget élvez,
  // különben a meglévő kép (ha van és nem törölték).
  const shownImage = previewUrl ?? (removeExisting ? null : existingImageUrl);

  const contentLength = content.trim().length;

  // Escape-re zárjon
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  // A háttér ne görögjön a modal alatt, és a fókusz rögtön az első mezőn legyen.
  useEffect(() => {
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    titleInputRef.current?.focus();
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, []);

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
      setError(
        "Nem támogatott képformátum. Engedélyezett: JPEG, PNG, WebP, GIF.",
      );
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
    // Új kép esetén a "meglévő törlése" jelző már lényegtelen (a feltöltés úgyis
    // felülírja), állítsuk vissza, hogy a submit logika egyértelmű legyen.
    setRemoveExisting(false);
    applyImage(file);
  };

  const clearImage = () => {
    applyImage(null);
    // Szerkesztéskor a meglévő képet is töröljük (remove_image=true a submitben).
    if (isEdit) setRemoveExisting(true);
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
    if (image) {
      formData.append("image", image);
    } else if (isEdit && removeExisting) {
      // Nincs új kép, de a meglévőt törölni kell.
      formData.append("remove_image", "true");
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
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/40 p-4 backdrop-blur-sm"
      onClick={onClose}
      role="presentation"
    >
      <div
        className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-panel border border-line bg-surface shadow-panel"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-post-title"
      >
        {/* Ragadós fejléc, hogy hosszú űrlapnál is elérhető maradjon a bezárás. */}
        <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-line bg-surface/95 px-6 py-4 backdrop-blur-sm">
          <h2 id="create-post-title" className="text-lg font-bold">
            {isEdit ? "Poszt szerkesztése" : "Új poszt létrehozása"}
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Bezárás"
            className="-mr-2 -mt-1 inline-flex h-8 w-8 items-center justify-center rounded-control text-xl leading-none text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground"
          >
            ×
          </button>
        </div>

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
              <>
                Markdown támogatott (címek, listák, táblázatok, **félkövér**…).{" "}
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
              </>
            }
          >
            <textarea
              id="post-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={9}
              placeholder={
                "A poszt tartalma...\n\n## Alcím\n**félkövér**, *dőlt*, [link](https://...)\n\n- felsorolás"
              }
              className={`${controlClass} resize-y font-mono`}
            />
          </Field>

          <Field label="Borítókép" htmlFor="post-image" optional>
            <input
              id="post-image"
              ref={fileInputRef}
              type="file"
              accept={ACCEPTED_IMAGE_TYPES.join(",")}
              onChange={handleImageChange}
              className="w-full cursor-pointer text-sm text-muted-foreground file:mr-3 file:cursor-pointer file:rounded-control file:border file:border-line file:bg-surface-muted file:px-4 file:py-2 file:text-sm file:font-medium file:text-foreground hover:file:bg-line"
            />
            {shownImage && (
              <div className="mt-3 flex items-start gap-3">
                {/* next/image helyett sima img: a kép a saját API-nkról jön, aminek
                    a domainje környezetenként más — így nem kell remotePatterns. */}
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={shownImage}
                  alt="Előnézet"
                  className="max-h-40 w-auto rounded-control border border-line"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={clearImage}
                >
                  Kép eltávolítása
                </Button>
              </div>
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
      </div>
    </div>
  );
}
