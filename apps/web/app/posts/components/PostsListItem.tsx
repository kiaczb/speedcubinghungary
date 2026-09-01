"use client";
import React, { useState } from "react";
import dynamic from "next/dynamic";
import Link from "next/link";
import Card from "@/app/components/Cards/Card";
import Badge from "@/app/components/Badge";
import Button from "@/app/components/ui/Button";
import { badgeMap, PostsItem } from "../types";
import { postImageUrl } from "../lib/postImage";
import { postPreview } from "../lib/postPreview";
import { reportSummary } from "../lib/report";

// Lásd CreatePostButton: a szerkesztő külön chunkban tölt be, csak ha kell.
const CreatePostModal = dynamic(() => import("./CreatePostModal"), {
  ssr: false,
});

type Props = {
  postsItem: PostsItem;
  checked: boolean;
  onToggle: () => void;
  isAdmin?: boolean;
};

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
        className={`group relative ${checked ? "border-foreground" : ""}`}
      >
        <div className="flex gap-4">
          {isAdmin && (
            <input
              type="checkbox"
              checked={checked}
              onChange={onToggle}
              aria-label={`${postsItem.title} kijelölése`}
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
              /* Sima img: a kép domainje környezetenként más — így nem kell remotePatterns. */
              /* eslint-disable-next-line @next/next/no-img-element */
              <img
                src={imageUrl}
                alt=""
                className="hidden h-24 w-32 shrink-0 rounded-control object-cover sm:block"
              />
            )}
          </div>

          {canEdit && (
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
