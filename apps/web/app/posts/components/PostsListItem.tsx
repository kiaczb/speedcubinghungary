"use client";
import React, { useState } from "react";
import Link from "next/link";
import Card from "@/app/components/Cards/Card";
import Badge from "@/app/components/Badge";
import Button from "@/app/components/ui/Button";
import { badgeMap, PostsItem } from "../types";
import { postImageUrl } from "../lib/postImage";
import { postPreview } from "../lib/postPreview";
import { reportSummary } from "../lib/report";
import CreatePostModal from "./CreatePostModal";

type Props = {
  postsItem: PostsItem;
  checked: boolean;
  onToggle: () => void;
  isAdmin?: boolean;
};

/**
 * Egy sor a hírlistában.
 *
 * Változások:
 * - a kijelölő "checkbox" korábban egy `div` volt `onClick`-kel és egy rejtett
 *   inputtal — billentyűzettel nem lehetett elérni. Most valódi `<input>`;
 * - a kártya `text-center`-t örökölt, ezért a cím és a kivonat középre zárt volt;
 * - a `font-display` és a `text-muted-foreground` osztályok nem is léteztek a
 *   témában, így a dátum szín nélkül maradt;
 * - a kép mostantól bélyegkép a szöveg mellett, nem egy 256 px magas fejléc,
 *   így egy képernyőre több hír fér ki.
 */
const PostsListItem = ({ postsItem, checked, onToggle, isAdmin }: Props) => {
  const badge = badgeMap[postsItem.tag];
  const imageUrl = postImageUrl(postsItem.image_path);
  const [editing, setEditing] = useState(false);
  // A riport tartalma JSON — összefoglalót mutatunk, nem a nyers tartalmat.
  const preview =
    postsItem.tag === "report"
      ? reportSummary(postsItem.content)
      : postPreview(postsItem.content);
  // A riport posztot NEM szerkesztjük ezzel a modallal: a tartalma JSON, a modal
  // viszont markdownt vár — a job írja/frissíti (idempotens).
  const canEdit = isAdmin && postsItem.tag !== "report";

  return (
    <>
      <Card
        interactive
        padding="md"
        // `relative` kell a cím kiterjesztett kattintófelületéhez (after:inset-0).
        className={`group relative ${checked ? "border-foreground" : ""}`}
      >
        <div className="flex gap-4">
          {isAdmin && (
            <input
              type="checkbox"
              checked={checked}
              onChange={onToggle}
              aria-label={`${postsItem.title} kijelölése`}
              // `relative z-10`: a cím kattintófelülete alá ne kerüljön.
              className="relative z-10 mt-1 h-4 w-4 shrink-0 cursor-pointer accent-foreground"
            />
          )}

          <div className="flex min-w-0 flex-1 gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5">
                <Badge color={badge.color} label={badge.label} />
                {postsItem.created_at && (
                  <time
                    dateTime={postsItem.created_at}
                    className="tabular font-mono text-xs text-subtle-foreground"
                  >
                    {new Date(postsItem.created_at).toLocaleDateString("hu-HU")}
                  </time>
                )}
              </div>

              <h2 className="mt-2 text-lg font-semibold leading-snug text-foreground">
                <Link
                  href={`/posts/${postsItem.id}`}
                  // A teljes sorra kiterjesztett kattintófelület, de a link
                  // maga a cím marad — így a képernyőolvasó is értelmes
                  // szöveget kap.
                  className="after:absolute after:inset-0 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-foreground"
                >
                  {postsItem.title}
                </Link>
              </h2>

              <p className="mt-1.5 line-clamp-2 text-sm leading-relaxed text-muted-foreground">
                {preview}
              </p>
            </div>

            {imageUrl && (
              /* next/image helyett sima img: a kép a saját API-nkról jön, aminek a
                 domainje környezetenként más — így nem kell remotePatterns config. */
              /* eslint-disable-next-line @next/next/no-img-element */
              <img
                src={imageUrl}
                alt=""
                className="hidden h-24 w-32 shrink-0 rounded-control object-cover sm:block"
              />
            )}
          </div>

          {canEdit && (
            // `relative z-10`, hogy a cím kiterjesztett kattintófelülete alá ne
            // kerüljön be a gomb.
            <div className="relative z-10 shrink-0">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setEditing(true)}
              >
                Szerkesztés
              </Button>
            </div>
          )}
        </div>
      </Card>

      {editing && (
        <CreatePostModal post={postsItem} onClose={() => setEditing(false)} />
      )}
    </>
  );
};

export default PostsListItem;
