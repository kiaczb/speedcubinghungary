"use client";
import { FaRegNewspaper } from "react-icons/fa";
import { PostsItem } from "../types";
import PostsListItem from "./PostsListItem";
import CreatePostButton from "./CreatePostButton";
import StateMessage from "@/app/components/ui/StateMessage";
import { useUser } from "@/app/components/UserProvider";

type Props = {
  posts: PostsItem[];
};

export default function PostsListClient({ posts }: Props) {
  const user = useUser();
  const isAdmin = user?.user?.roles.includes("admin");

  return (
    <>
      {isAdmin && (
        <div className="mb-6 flex flex-wrap items-center gap-3 rounded-card border border-line bg-surface p-3">
          <CreatePostButton />
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
            <PostsListItem key={post.id} postsItem={post} isAdmin={isAdmin} />
          ))}
        </div>
      )}
    </>
  );
}
