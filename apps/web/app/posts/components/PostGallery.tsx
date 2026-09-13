"use client";
import { useRef, useState } from "react";
import { FaChevronLeft, FaChevronRight } from "react-icons/fa";

type Props = {
  /** A képek teljes URL-jei a megjelenítés sorrendjében (lásd postImageUrls). */
  images: string[];
  /** A poszt címe — a képek alt szövegéhez, hogy ne üres alt maradjon. */
  title: string;
};

// Ennyi vízszintes elhúzás számít lapozásnak (px). Kisebb mozdulat véletlen
// érintés is lehet, és a függőleges görgetést nem akarjuk elnyelni.
const SWIPE_THRESHOLD = 50;

/**
 * A poszt képgalériája. Egyetlen képnél nincs vezérlő — akkor ez pontosan az a
 * kép, ami korábban is látszott a poszt oldalán; több képnél nyilakkal, pöttyökkel,
 * billentyűzettel és mobilon húzással lehet lapozni. A lapozás KÖRBEÉR: az utolsó
 * kép után újra az első jön.
 */
export default function PostGallery({ images, title }: Props) {
  const [index, setIndex] = useState(0);
  // Az érintés kezdőpontja — ref, mert a húzás közben nem kell újrarenderelni.
  const touchStartX = useRef<number | null>(null);

  if (images.length === 0) return null;

  if (images.length === 1) {
    return (
      /* Sima img: a kép domainje környezetenként más — így nem kell remotePatterns. */
      /* eslint-disable-next-line @next/next/no-img-element */
      <img
        src={images[0]}
        alt={title}
        className="mt-8 w-full rounded-panel border border-line object-cover"
      />
    );
  }

  const go = (delta: number) =>
    setIndex((current) => (current + delta + images.length) % images.length);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "ArrowLeft") {
      e.preventDefault();
      go(-1);
    } else if (e.key === "ArrowRight") {
      e.preventDefault();
      go(1);
    }
  };

  const handleTouchEnd = (e: React.TouchEvent<HTMLDivElement>) => {
    const start = touchStartX.current;
    touchStartX.current = null;
    if (start === null) return;

    const distance = start - e.changedTouches[0].clientX;
    if (Math.abs(distance) < SWIPE_THRESHOLD) return;
    go(distance > 0 ? 1 : -1);
  };

  return (
    <section
      className="mt-8"
      aria-roledescription="carousel"
      aria-label={`${title} — képek`}
    >
      <div
        tabIndex={0}
        onKeyDown={handleKeyDown}
        onTouchStart={(e) => {
          touchStartX.current = e.touches[0].clientX;
        }}
        onTouchEnd={handleTouchEnd}
        className="group relative overflow-hidden rounded-panel border border-line bg-surface-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
      >
        {/* Egy sáv, amit eltolunk — így a váltás animált, és minden kép a DOM-ban
            marad (a böngésző előre betölti a következőt). */}
        <div
          className="flex transition-transform duration-500 ease-smooth"
          style={{ transform: `translateX(-${index * 100}%)` }}
        >
          {images.map((src, i) => (
            <div key={src} className="w-full shrink-0">
              {/* Sima img: a kép domainje környezetenként más — így nem kell remotePatterns. */}
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={src}
                alt={`${title} — ${i + 1}. kép`}
                // A nem látszó képeket kivesszük a képernyőolvasó és a
                // tabsorrend elől, hogy a galéria egy elemként viselkedjen.
                aria-hidden={i !== index}
                className="aspect-video w-full object-cover"
              />
            </div>
          ))}
        </div>

        <button
          type="button"
          onClick={() => go(-1)}
          aria-label="Előző kép"
          className="absolute left-3 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-surface/85 text-foreground shadow-card backdrop-blur-sm transition-colors duration-200 ease-smooth hover:bg-surface"
        >
          <FaChevronLeft className="h-3.5 w-3.5" />
        </button>
        <button
          type="button"
          onClick={() => go(1)}
          aria-label="Következő kép"
          className="absolute right-3 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-surface/85 text-foreground shadow-card backdrop-blur-sm transition-colors duration-200 ease-smooth hover:bg-surface"
        >
          <FaChevronRight className="h-3.5 w-3.5" />
        </button>

        <p
          aria-live="polite"
          className="tabular absolute right-3 top-3 rounded-control bg-surface-inverse/70 px-2 py-1 font-mono text-xs text-white backdrop-blur-sm"
        >
          {index + 1} / {images.length}
        </p>
      </div>

      <div className="mt-3 flex flex-wrap items-center justify-center gap-2">
        {images.map((src, i) => (
          <button
            key={src}
            type="button"
            onClick={() => setIndex(i)}
            aria-label={`${i + 1}. kép megjelenítése`}
            aria-current={i === index}
            className={`h-2 rounded-full transition-all duration-200 ease-smooth ${
              i === index
                ? "w-6 bg-foreground"
                : "w-2 bg-line-strong hover:bg-muted-foreground"
            }`}
          />
        ))}
      </div>
    </section>
  );
}
