"use client";
import React, { useState } from "react";
import dynamic from "next/dynamic";
import Link from "next/link";
import { FaRegEdit } from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import Badge from "@/app/components/Badge";
import { badgeMap, PostsItem } from "../types";
import { postCoverUrl } from "../lib/postImage";
import { postPreview } from "../lib/postPreview";
import { formatShortDate } from "@/app/lib/date";
import { reportSummary } from "../lib/report";
import DeletePostButton from "./DeletePostButton";

// Lásd CreatePostButton: a szerkesztő külön chunkban tölt be, csak ha kell.
const CreatePostModal = dynamic(() => import("./CreatePostModal"), {
  ssr: false,
});

type Props = {
  postsItem: PostsItem;
  isAdmin?: boolean;
};

const PostsListItem = ({ postsItem, isAdmin }: Props) => {
  const badge = badgeMap[postsItem.tag];
  const imageUrl = postCoverUrl(postsItem.image_paths);
  // A borítókép mellett jelezzük, ha a poszthoz több kép is tartozik.
  const imageCount = postsItem.image_paths?.length ?? 0;
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
      <Card interactive padding="md" className="group relative">
        <div className="flex gap-4">
          <div className="flex min-w-0 flex-1 gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5">
                <Badge color={badge.color} label={badge.label} />
                {postsItem.created_at && (
                  <time
                    dateTime={postsItem.created_at}
                    className="tabular font-mono text-xs text-subtle-foreground"
                  >
                    {formatShortDate(postsItem.created_at)}
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
              <div className="relative hidden h-24 w-32 shrink-0 sm:block">
                <img
                  src={imageUrl}
                  alt=""
                  className="h-full w-full rounded-control object-cover"
                />
                {imageCount > 1 && (
                  <span className="tabular absolute bottom-1 right-1 rounded-control bg-surface-inverse/75 px-1.5 py-0.5 font-mono text-[11px] text-white">
                    {imageCount} kép
                  </span>
                )}
              </div>
            )}
          </div>

          {isAdmin && (
            <div className="relative z-10 flex shrink-0 items-center gap-1">
              {canEdit && (
                <button
                  type="button"
                  onClick={() => setEditing(true)}
                  aria-label={`${postsItem.title} szerkesztése`}
                  className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground"
                >
                  <FaRegEdit className="h-4 w-4" />
                </button>
              )}
              <DeletePostButton postId={postsItem.id} title={postsItem.title} />
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
