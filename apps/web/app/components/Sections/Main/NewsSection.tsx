import NewsCard from "../../Cards/NewsCard";
import Section from "./Section";
import Button from "../../ui/Button";
import { FaArrowRight } from "react-icons/fa";
import { PostsItem } from "@/app/posts/types";
import { postCoverUrl } from "@/app/posts/lib/postImage";
import { serverJson } from "@/app/lib/api";

async function getLatestPosts(): Promise<PostsItem[]> {
  return serverJson<PostsItem[]>("/posts/latest", []);
}

export default async function NewsSection() {
  const latestPosts = await getLatestPosts();

  if (latestPosts.length === 0) return null;

  return (
    <Section bleed className="bg-surface-inverse py-20 sm:py-24">
      <div className="mx-auto w-full max-w-6xl px-5 sm:px-8 lg:px-12">
        <div className="max-w-2xl">
          <p className="mb-3 font-mono text-xs font-medium uppercase tracking-[0.18em] text-white/40">
            Friss
          </p>
          <h2 className="text-3xl font-bold text-white sm:text-4xl">Hírek</h2>
          <p className="mt-4 text-base leading-relaxed text-white/60 sm:text-lg">
            Versenyek, események és eredmények a magyar kockázó közösségből.
          </p>
        </div>

        <div className="mt-12 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {latestPosts.map((post) => (
            <NewsCard
              key={post.id}
              post={post}
              imageUrl={postCoverUrl(post.image_paths)}
            />
          ))}
        </div>

        <div className="mt-12">
          <Button
            href="/posts"
            variant="inverse"
            size="lg"
            iconRight={<FaArrowRight className="h-3.5 w-3.5" />}
          >
            Összes hír megtekintése
          </Button>
        </div>
      </div>
    </Section>
  );
}
