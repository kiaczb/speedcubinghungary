"use client";
import { useState } from "react";
import { FaRegNewspaper } from "react-icons/fa";
import { PostsItem } from "../types";
import PostsListItem from "./PostsListItem";
import DeletePostsButton from "./DeletePostsButton";
import CreatePostButton from "./CreatePostButton";
import StateMessage from "@/app/components/ui/StateMessage";
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
        <div className="mb-6 flex flex-wrap items-center gap-3 rounded-card border border-line bg-surface p-3">
          <CreatePostButton />
          <DeletePostsButton
            selectedIds={selectedIds}
            onDeleted={() => setSelectedIds([])}
          />
          {selectedIds.length > 0 && (
            <span className="ml-auto font-mono text-xs text-muted-foreground">
              {selectedIds.length} kijelölve
            </span>
          )}
        </div>
      )}

      {posts.length === 0 ? (
        <StateMessage
          icon={<FaRegNewspaper className="h-5 w-5" />}
          title="Még nincs közzétett hír"
          description="Amint megjelenik az első bejegyzés, itt fogod látni."
        />
      ) : (
        <div className="flex flex-col gap-4">
          {posts.map((post) => (
            <PostsListItem
              key={post.id}
              postsItem={post}
              checked={selectedIds.includes(post.id)}
              onToggle={() => toggle(post.id)}
              isAdmin={isAdmin}
            />
          ))}
        </div>
      )}
    </>
  );
}
