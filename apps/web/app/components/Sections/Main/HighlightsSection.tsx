import { Suspense } from "react";
import Section from "./Section";
import { formatLongDateWithWeekday } from "@/app/lib/date";
import AllRounderTile from "./Highlights/AllRounderTile";
import FeaturedTile from "./Highlights/FeaturedTile";
import UpcomingCompetitionsTile from "./Highlights/UpcomingCompetitionsTile";
import { TileSkeleton } from "./Highlights/Tile";

const FEATURED = "sm:col-span-2 lg:row-span-2";

const HighlightsSection = () => {
  const today = new Date();

  return (
    <Section>
      <div className="mb-6 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1 border-b border-line pb-3">
        <h2 className="text-2xl font-bold text-foreground">Aktuális</h2>
        <time
          dateTime={today.toISOString()}
          className="tabular font-mono text-sm text-muted-foreground"
        >
          {formatLongDateWithWeekday(today)}
        </time>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Suspense fallback={<TileSkeleton className={FEATURED} lines={6} />}>
          <AllRounderTile className={FEATURED} />
        </Suspense>
        <Suspense fallback={<TileSkeleton lines={3} />}>
          <FeaturedTile />
        </Suspense>
        <Suspense fallback={<TileSkeleton lines={4} />}>
          <UpcomingCompetitionsTile />
        </Suspense>
      </div>
    </Section>
  );
};

export default HighlightsSection;
