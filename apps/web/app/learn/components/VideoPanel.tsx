import { FaExternalLinkAlt } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import type { VideoEmbed } from "../lib/video";

type Props = {
  video: VideoEmbed;
  title: string;
};

// A hívó felelőssége, hogy videó nélkül NE rendereljen üres keretet — ezért ez
// a komponens már csak létező `VideoEmbed`-et kap.
export default function VideoPanel({ video, title }: Props) {
  if (video.kind === "link") {
    return (
      <div className="flex flex-col items-start gap-4 rounded-card border border-line bg-surface p-6">
        <p className="text-sm leading-relaxed text-muted-foreground">
          Ehhez az algoritmushoz videó tartozik, de a szolgáltatót nem tudjuk
          beágyazni. Nyisd meg új lapon.
        </p>
        <Button
          href={video.url}
          external
          variant="secondary"
          size="sm"
          icon={<FaExternalLinkAlt className="h-3 w-3" />}
        >
          Videó megnyitása
        </Button>
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-card border border-line bg-surface-inverse shadow-card">
      {/* 16:9 arány — a videó a saját szélességéhez igazodik, így mobilon
          ugyanúgy kitölti az oszlopot, mint asztali gépen. */}
      <div className="relative aspect-video w-full">
        {video.kind === "youtube" ? (
          <iframe
            src={video.src}
            title={`${title} — videó`}
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
            allowFullScreen
            loading="lazy"
            referrerPolicy="strict-origin-when-cross-origin"
            className="absolute inset-0 h-full w-full border-0"
          />
        ) : (
          <video
            src={video.src}
            controls
            preload="metadata"
            className="absolute inset-0 h-full w-full"
          />
        )}
      </div>
    </div>
  );
}
