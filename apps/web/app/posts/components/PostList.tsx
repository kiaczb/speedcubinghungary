"use client";
import { useState } from "react";
import { PostsItem } from "../types";
import PostsListItem from "./PostsListItem";
import DeletePostsButton from "./DeletePostsButton";
import CreatePostButton from "./CreatePostButton";
import { useUser } from "@/app/components/UserProvider";

type Props = {
  posts: PostsItem[];
};

export default function PostsListClient({ posts }: Props) {
  const user = useUser();
  const isAdmin = user?.user?.roles.includes("admin");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  const toggle = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  };

  return (
    <>
      {isAdmin && (
        <div>
          <CreatePostButton />

          <DeletePostsButton
            selectedIds={selectedIds}
            onDeleted={() => setSelectedIds([])}
          />
        </div>
      )}

      <div className="space-y-6">
        {posts.map((post) => (
          <PostsListItem
            key={post.id}
            postsItem={post}
            checked={selectedIds.includes(post.id)}
            onToggle={() => toggle(post.id)}
          />
        ))}
      </div>
    </>
  );
}
