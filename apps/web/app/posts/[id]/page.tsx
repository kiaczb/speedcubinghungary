import React from "react";
import { PostsItem } from "../types";
import { notFound } from "next/dist/client/components/navigation";
import { postImageUrl } from "../lib/postImage";

type PageProps = {
  params: Promise<{
    id: string;
  }>;
};

async function getPost(id: string): Promise<PostsItem | null> {
  const baseUrl = process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const res = await fetch(`${baseUrl}/posts/${id}`, { 
    cache: "no-store" 
  });

  if (!res.ok) return null;
  return res.json();
}

export default async function PostPage({ params }: PageProps) {
  const { id } = await params;
  const post = await getPost(id);
  if (!post) {
    notFound(); // 404 oldal
  }

  const imageUrl = postImageUrl(post.image_path);

  return (
    <article className="max-w-3xl mx-auto py-10">
      <h1 className="text-3xl font-bold mb-4">{post.title}</h1>
      {imageUrl && (
        /* next/image helyett sima img: a kép a saját API-nkról jön, aminek a
           domainje környezetenként más — így nem kell remotePatterns config. */
        /* eslint-disable-next-line @next/next/no-img-element */
        <img
          src={imageUrl}
          alt={post.title}
          className="mb-6 w-full rounded-lg object-cover"
        />
      )}
      <div className="prose">{post.content}</div>
    </article>
  );
}
