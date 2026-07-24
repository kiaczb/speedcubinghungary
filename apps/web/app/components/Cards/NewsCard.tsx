import React from "react";
import Card from "./Card";
import Link from "next/link";
import Badge from "../Badge";
import { badgeMap, PostsItem, TagTypes } from "@/app/posts/types";
import { postPreview } from "@/app/posts/lib/postPreview";
import { reportSummary } from "@/app/posts/lib/report";

// Dictionary a típusokhoz

type NewsCardProps = {
  type: TagTypes;
  imageUrl?: string;
  post: PostsItem;
};

const NewsCard: React.FC<NewsCardProps> = ({ type, imageUrl, post }) => {
  const badge = badgeMap[type];
  const preview =
    post.tag === "report"
      ? reportSummary(post.content)
      : postPreview(post.content, 120);
  return (
    <Card className="max-w-86 flex flex-col justify-between bg-white p-0 overflow-hidden hover:shadow-2xl transition-all duration-300 hover:-translate-y-1">
      {/* Header rész az új dizájnnal */}
      <div>
        <div className="relative">
          <Badge
            label={badge.label}
            color={badge.color}
            className="absolute top-2 left-3 z-10"
          />

          {imageUrl && (
            /* next/image helyett sima img: a kép lehet a saját API-nkról (posztkép,
               környezetenként más domain) vagy a public mappából — mindkettőt
               remotePatterns nélkül ki kell tudni szolgálni. */
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={imageUrl}
              alt={post.title}width={350}
    height={100}
    className="max-w-full object-cover"
            />
          )}
        </div>
      </div>

      <div className="p-6 pb-4 mt-2">
        <div>
          <h1 className="text-xl font-bold text-gray-900 mb-3 group-hover:text-blue-600 transition-colors">
            {post.title}
          </h1>

          {/* Dátum és olvasási idő */}
          <div className="flex items-center justify-center text-sm text-gray-500 mb-3">
            <span className="flex items-center mr-4">
              <span className="mr-1">📅</span>
              <span>
                {new Date(post.created_at).toLocaleDateString("hu-HU")}
              </span>
            </span>
          </div>
        </div>
        {/* <Badge label={badge.label} color={badge.color} /> */}
      </div>
      {/* Tartalom rész */}
      <div className="px-6 pb-6 flex flex-1 flex-col">
        <div className="text-md text-gray-700 mb-6">
          <p className="leading-relaxed">{preview}</p>
        </div>

        {/* Link rész - modernizálva */}
        <div className="flex justify-center pt-4 border-t border-gray-100 mt-auto text-sm">
          <Link
            href="/about"
            className="group inline-flex items-center px-6 py-3 bg-black text-white font-semibold rounded-lg border border-gray-800 hover:border-gray-200 duration-300 shadow-lg hover:shadow-xl"
          >
            <span>Részletek</span>
          </Link>
        </div>
      </div>
    </Card>
  );
};

export default NewsCard;
