import React from "react";
import Link from "next/link";
import Card from "./Card";
import Badge from "../Badge";
import { badgeMap, PostsItem } from "@/app/posts/types";
import { postPreview } from "@/app/posts/lib/postPreview";
import { reportSummary } from "@/app/posts/lib/report";
import { FaArrowRight } from "react-icons/fa";

type NewsCardProps = {
  post: PostsItem;
  imageUrl?: string | null;
};

const NewsCard: React.FC<NewsCardProps> = ({ post, imageUrl }) => {
  const badge = badgeMap[post.tag];
  const preview =
    post.tag === "report"
      ? reportSummary(post.content)
      : postPreview(post.content, 120);

  return (
    <Card interactive className="group overflow-hidden">
      <Link
        href={`/posts/${post.id}`}
        className="flex h-full flex-col rounded-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
      >
        <div className="relative aspect-video overflow-hidden bg-surface-muted">
          {imageUrl ? (
            /* Sima img: a kép domainje környezetenként más — így nem kell remotePatterns. */
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={imageUrl}
              alt=""
              className="h-full w-full object-cover transition-transform duration-500 ease-smooth group-hover:scale-[1.03]"
            />
          ) : (
            <div className="h-full w-full bg-linear-to-br from-surface-muted to-line" />
          )}
          <Badge
            label={badge.label}
            color={badge.color}
            className="absolute left-3 top-3 bg-white/90 backdrop-blur-sm"
          />
        </div>

        <div className="flex flex-1 flex-col p-5">
          <time
            dateTime={post.created_at}
            className="tabular font-mono text-xs text-subtle-foreground"
          >
            {new Date(post.created_at).toLocaleDateString("hu-HU")}
          </time>

          <h3 className="mt-2 text-lg font-semibold leading-snug text-foreground">
            {post.title}
          </h3>

          <p className="mt-2 line-clamp-3 text-sm leading-relaxed text-muted-foreground">
            {preview}
          </p>

          <span className="mt-5 inline-flex items-center gap-2 border-t border-line pt-4 text-sm font-semibold text-foreground">
            Részletek
            <FaArrowRight className="h-3 w-3 transition-transform duration-200 ease-smooth group-hover:translate-x-1" />
          </span>
        </div>
      </Link>
    </Card>
  );
};

export default NewsCard;
