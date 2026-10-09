import type { Metadata } from "next";
import Link from "next/link";
import "@cubing/icons";
import { FaArrowRight } from "react-icons/fa";
import Card from "../components/Cards/Card";
import Button from "../components/ui/Button";
import PageHeader from "../components/ui/PageHeader";
import { serverJson } from "../lib/api";
import VideoPlaceholder from "./components/VideoPlaceholder";
import { ALGORITHMS_DESCRIPTION, ALGORITHMS_HREF, levelApiPath, levelHref } from "./lib/nodes";
import type { LearnLevel } from "./types";

const DESCRIPTION =
  "Videók a kezdéshez és az első versenyhez, valamint egy rendezett algoritmusgyűjtemény — kirakónként, módszerenként, esetenként.";

export const metadata: Metadata = {
  title: "Tanulás",
  description: DESCRIPTION,
};

const VIDEOS = [
  {
    id: "intro",
    eyebrow: "Bemutató",
    title: "Mi az a speedcubing?",
    description:
      "Rövid bevezető a kockázás világába: honnan indul, mennyire lehet gyors, és miért éri meg belevágni.",
  },
  {
    id: "competitor-guide",
    eyebrow: "Versenyzőknek",
    title: "Versenyzői útmutató",
    description:
      "Mi történik egy WCA versenyen? Nevezés, a kirakások menete, a bírók és a keverők szerepe — lépésről lépésre.",
  },
];

export default async function LearnPage() {
  const root = await serverJson<LearnLevel | null>(levelApiPath([]), null);
  const topics = root?.children ?? [];

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader title="Tanulás" description={DESCRIPTION} />

        <div className="mt-12 grid grid-cols-1 gap-6 lg:grid-cols-2">
          {VIDEOS.map((video) => (
            <Card key={video.id} className="flex flex-col overflow-hidden">
              <div className="p-3 pb-0">
                <VideoPlaceholder label={video.title} />
              </div>
              <div className="flex flex-1 flex-col gap-2 p-5 sm:p-6">
                <p className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-subtle-foreground">
                  {video.eyebrow}
                </p>
                <h2 className="text-xl font-bold text-foreground">{video.title}</h2>
                <p className="text-sm leading-relaxed text-muted-foreground">
                  {video.description}
                </p>
              </div>
            </Card>
          ))}
        </div>

        <section className="mt-14 rounded-panel bg-surface-inverse px-6 py-10 sm:px-10 sm:py-12">
          <p className="mb-3 font-mono text-xs font-medium uppercase tracking-[0.18em] text-white/40">
            Algoritmusok
          </p>
          <h2 className="max-w-xl text-2xl font-bold text-white sm:text-3xl">
            Tanuld meg az algoritmusokat, lépésről lépésre
          </h2>
          <p className="mt-3 max-w-xl text-base leading-relaxed text-white/60">
            {ALGORITHMS_DESCRIPTION}
          </p>

          {topics.length > 0 && (
            <ul className="mt-6 flex flex-wrap gap-2">
              {topics.map((topic) => (
                <li key={topic.id}>
                  <Link
                    href={levelHref(topic.path)}
                    className="inline-flex items-center gap-2 rounded-control border border-white/15 bg-white/5 px-3 py-1.5 text-sm font-medium text-white/80 transition-colors duration-200 ease-smooth hover:border-white/30 hover:bg-white/10 hover:text-white focus-visible:outline-white"
                  >
                    {topic.icon && (
                      <span
                        className={`cubing-icon ${topic.icon} text-base leading-none`}
                        aria-hidden
                      />
                    )}
                    {topic.title}
                  </Link>
                </li>
              ))}
            </ul>
          )}

          <div className="mt-8">
            <Button
              href={ALGORITHMS_HREF}
              variant="inverse"
              size="lg"
              iconRight={<FaArrowRight className="h-3.5 w-3.5" />}
            >
              Irány az algoritmusok
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
