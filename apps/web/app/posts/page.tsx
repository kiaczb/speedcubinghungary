import type { Metadata } from "next";
import PostsList from "./components/PostList";
import PageHeader from "../components/ui/PageHeader";
import { serverJson } from "../lib/api";
import { PostsItem } from "./types";

export const metadata: Metadata = {
  title: "Hírek",
  description:
    "Versenyek, események és heti riportok a magyar kockázó közösségből.",
};

async function getPosts(): Promise<PostsItem[]> {
  return serverJson<PostsItem[]>("/posts", []);
}

export default async function PostsPage() {
  const posts = await getPosts();

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <PageHeader
          title="Hírek"
          description="Itt találod az összes közzétett hírünket, versenybeszámolónkat és heti riportunkat."
        />

        <div className="mt-10">
          <PostsList posts={posts} />
        </div>
      </div>
    </div>
  );
}
