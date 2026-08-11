import React from "react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { FaArrowLeft } from "react-icons/fa";
import { badgeMap, PostsItem } from "../types";
import Badge from "@/app/components/Badge";
import { serverApiUrl } from "@/app/lib/api";
import { postImageUrl } from "../lib/postImage";
import PostContent from "../components/PostContent";
import WeeklyReportView from "../components/WeeklyReportView";
import { parseWeeklyReport } from "../lib/report";

type PageProps = {
  params: Promise<{
    id: string;
  }>;
};

async function getPost(id: string): Promise<PostsItem | null> {
  const res = await fetch(`${serverApiUrl()}/posts/${id}`, {
    cache: "no-store",
  });

  if (!res.ok) return null;
  return res.json();
}

export default async function PostPage({ params }: PageProps) {
  const { id } = await params;
  const post = await getPost(id);
  if (!post) {
    notFound();
  }

  const imageUrl = postImageUrl(post.image_path);
  const badge = badgeMap[post.tag];

  // A riport posztok tartalma strukturált JSON — dedikált komponens rendereli.
  // Törött/régi JSON esetén a markdown-renderelőre esik vissza, hogy sose
  // dőljön el az oldal.
  const report = post.tag === "report" ? parseWeeklyReport(post.content) : null;

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <article className="mx-auto w-full max-w-3xl">
        <Link
          href="/posts"
          className="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
        >
          <FaArrowLeft className="h-3 w-3 transition-transform duration-200 ease-smooth group-hover:-translate-x-0.5" />
          Vissza a hírekhez
        </Link>

        <header className="mt-6 border-b border-line pb-8">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
            <Badge color={badge.color} label={badge.label} />
            {post.created_at && (
              <time
                dateTime={post.created_at}
                className="tabular font-mono text-xs text-subtle-foreground"
              >
                {new Date(post.created_at).toLocaleDateString("hu-HU", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </time>
            )}
          </div>

          <h1 className="mt-4 text-3xl font-bold leading-tight text-foreground sm:text-4xl">
            {post.title}
          </h1>
        </header>

        {imageUrl && (
          /* Sima img: a kép domainje környezetenként más — így nem kell remotePatterns. */
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={imageUrl}
            alt=""
            className="mt-8 w-full rounded-panel border border-line object-cover"
          />
        )}

        <div className="mt-8 text-[15px] leading-relaxed text-foreground">
          {report ? (
            <WeeklyReportView report={report} />
          ) : (
            <PostContent content={post.content} />
          )}
        </div>
      </article>
    </div>
  );
}
