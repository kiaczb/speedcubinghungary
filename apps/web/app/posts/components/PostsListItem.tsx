import React from "react";
import Card from "@/app/components/Cards/Card";
import Link from "next/link";
import { badgeMap, PostsItem } from "../types";
import Badge from "@/app/components/Badge";
import { postImageUrl } from "../lib/postImage";
import { postPreview } from "../lib/postPreview";
import { reportSummary } from "../lib/report";

type Props = {
  postsItem: PostsItem;
  checked: boolean;
  onToggle: () => void;
};

const PostsListItem = ({ postsItem, checked, onToggle }: Props) => {
  const badge = badgeMap[postsItem.tag];
  const imageUrl = postImageUrl(postsItem.image_path);
  // A riport tartalma JSON — összefoglalót mutatunk, nem a nyers tartalmat.
  const preview =
    postsItem.tag === "report"
      ? reportSummary(postsItem.content)
      : postPreview(postsItem.content);
  return (
    <div>
      <Card className="p-5 rounded-lg">
        <div
          onClick={onToggle}
          className={`w-5 h-5 flex items-center justify-center self-start cursor-pointer ${
            checked ? "bg-green-500" : "bg-gray-200"
          }`}
        >
          <input
            type="checkbox"
            checked={checked}
            readOnly
            className="hidden"
          />
          {checked && <span className="text-white text-xl">✔</span>}
        </div>
        <Link href={`/posts/${postsItem.id}`} className="block">
          {imageUrl && (
            /* next/image helyett sima img: a kép a saját API-nkról jön, aminek a
               domainje környezetenként más — így nem kell remotePatterns config. */
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={imageUrl}
              alt={postsItem.title}
              className="mb-3 max-h-64 w-full rounded-md object-cover"
            />
          )}
          <Badge color={badge.color} label={badge.label} />
          <h3 className="font-display text-lg font-semibold mb-2">
            {postsItem.title}
          </h3>
          <p className="text-sm text-foreground mb-3">{preview}</p>
          {postsItem.created_at && (
            <span className="text-xs text-muted-foreground">
              {new Date(postsItem.created_at).toLocaleDateString()}
            </span>
          )}
        </Link>
      </Card>
    </div>
  );
};

export default PostsListItem;
