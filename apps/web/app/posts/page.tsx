import CreatePostButton from "./components/CreatePostButton";
import PostsList from "./components/PostList";
import PostsListItem from "./components/PostsListItem";
import { PostsItem } from "./types";

// Server oldali fetch
async function getPosts(): Promise<PostsItem[]> {
  // Docker környezetben az 'api' a szolgáltatás neve a docker-compose-ban
  const baseUrl = process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  try {
    const res = await fetch(`${baseUrl}/posts`, { cache: 'no-store' });
    if (!res.ok) return [];
    return res.json();
  } catch (error) {
    console.error("Nem sikerült elérni a backendet:", error);
    return [];
  }
}

export default async function PostsPage() {
  const posts = await getPosts();
  return (
    <div className="min-h-screen py-8 px-4">
      <div className="max-w-4xl mx-auto">
        {/* Fejléc */}
        <div className="mb-10">
          <h1 className="text-4xl font-bold mb-3">Összes hír</h1>
          <p className="text-gray-600">
            Itt találod az összes közzétett hírünket.
          </p>
        </div>

        <div>
          <PostsList posts={posts} />
        </div>
      </div>
    </div>
  );
}
