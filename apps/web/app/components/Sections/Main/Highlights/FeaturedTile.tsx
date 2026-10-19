import { serverJson } from "@/app/lib/api";
import { parseWeeklyReport } from "@/app/posts/lib/report";
import { PostsItem } from "@/app/posts/types";
import { PredictionContest } from "@/app/tools/prediction/types";
import {
  OpenPredictionTile,
  PredictionFallbackTile,
  latestContest,
  nearestOpenContest,
} from "./PredictionTile";
import { ReportTile } from "./ReportTile";

function latestReportPost(posts: PostsItem[]): PostsItem | undefined {
  return posts
    .filter((post) => post.tag === "report")
    .sort((a, b) => b.created_at.localeCompare(a.created_at))[0];
}

export default async function FeaturedTile({
  className = "",
}: {
  className?: string;
}) {
  const [contests, posts] = await Promise.all([
    serverJson<PredictionContest[]>("/predictions", []),
    serverJson<PostsItem[]>("/posts", []),
  ]);

  const open = nearestOpenContest(contests);
  if (open) {
    return (
      <OpenPredictionTile
        contest={open}
        now={new Date()}
        className={className}
      />
    );
  }

  const reportPost = latestReportPost(posts);
  const report = reportPost ? parseWeeklyReport(reportPost.content) : null;
  if (reportPost && report) {
    return (
      <ReportTile post={reportPost} report={report} className={className} />
    );
  }

  return (
    <PredictionFallbackTile
      latest={latestContest(contests)}
      className={className}
    />
  );
}
