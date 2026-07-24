import NewsCard from "../../Cards/NewsCard";
import Link from "next/link";
import Section from "./Section";
import { FaArrowRight } from "react-icons/fa";
import { PostsItem } from "@/app/posts/types";
import { postImageUrl } from "@/app/posts/lib/postImage";

async function getLatestPosts(): Promise<PostsItem[]> {
  const baseUrl = process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  try {
    const res = await fetch(`${baseUrl}/posts/latest`, { cache: 'no-store' });
    if (!res.ok) return [];
    return res.json();
  } catch (error) {
    console.error("Nem sikerült elérni a backendet:", error);
    return [];
  }
}

export default async function NewsSection() {
  const latestPosts = await getLatestPosts();
  return (
    <Section className="relative overflow-hidden">
      {/* Dizájn elemek */}
      <div className="absolute inset-0 bg-linear-to-br from-gray-900 via-gray-800 to-gray-900"></div>

      {/* Grid pattern overlay */}
      <div className="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.05)_1px,transparent_1px)] bg-[size-50px_50px] opacity-30"></div>

      <div className="relative z-10">
        <div className="container mx-auto px-4 py-16">
          {/* Fejléc */}
          <div className="text-center mb-12">
            <h1 className="text-4xl md:text-5xl font-bold mb-6 text-white">
              Hírek
            </h1>
            <p className="text-lg text-gray-300 max-w-3xl mx-auto leading-relaxed">
              Legfrissebb híreink, versenyek és események a speedcubing
              közösségből. Maradj naprakész mindennel, ami a magyar kockázó
              világban történik!
            </p>
          </div>

          {/* Kártyák - egyenlő méretűek, flex wrap */}
          <div className="flex flex-wrap justify-around gap-14 ">
            {latestPosts.map((post) => (
              <NewsCard
                key={post.id}
                type={post.tag}
                post={post}
                imageUrl={postImageUrl(post.image_path) ?? "/wanna_start_cubing.jpg"}
              />
            ))}
          </div>

          {/* Gomb */}
          <div className="text-center mt-20">
            <Link
              href="/posts"
              className="group inline-flex items-center px-4 py-2 lg:px-6 lg:py-4 bg-white hover:bg-gray-50 text-gray-900 font-semibold rounded-xl transition-all duration-300 shadow-lg hover:shadow-2xl hover:-translate-y-1 border border-gray-800 hover:border-gray-600"
            >
              <span className="text-lg font-bold tracking-tight">
                Összes hír megtekintése
              </span>
              <FaArrowRight className="w-5 h-5 ml-4 transform group-hover:translate-x-2 transition-transform duration-300" />
            </Link>
          </div>
        </div>
      </div>
    </Section>
  );
}
