import { parseVideo } from "../lib/video";
import type { LearnAlgorithmDetail } from "../types";
import AlgorithmPager from "./AlgorithmPager";
import CopyNotationButton from "./CopyNotationButton";
import VideoPanel from "./VideoPanel";

type Props = {
  detail: LearnAlgorithmDetail;
};

// A tanulónézet: az algoritmus és a videó asztali gépen EGYMÁS MELLETT, mobilon
// egymás alatt. Videó nélkül nincs üres keret — ilyenkor az algoritmus a teljes
// szélességet kapja.
export default function AlgorithmStudy({ detail }: Props) {
  const { algorithm, case: learnCase, siblings } = detail;
  const video = parseVideo(algorithm.videoUrl);

  return (
    <>
      <div
        className={`grid gap-6 ${video ? "lg:grid-cols-2 lg:items-start" : ""}`}
      >
        <div className="flex flex-col gap-6">
          <section
            aria-label="Algoritmus"
            className="rounded-card border border-line bg-surface shadow-card"
          >
            <div className="flex items-center justify-between gap-4 border-b border-line px-5 py-3">
              <h2 className="text-sm font-semibold text-foreground">
                Algoritmus
              </h2>
              <CopyNotationButton notation={algorithm.notation} />
            </div>

            {/* A notáció a nézet főszereplője: nagy, monospace, tördelhető —
                egy hosszabb algoritmus telefonon is elolvasható maradjon. */}
            <p className="tabular break-words px-5 py-8 text-center font-mono text-2xl leading-relaxed text-foreground sm:text-3xl">
              {algorithm.notation}
            </p>
          </section>

          {algorithm.description && (
            <section
              aria-label="Leírás"
              className="rounded-card border border-line bg-surface p-5"
            >
              <h2 className="text-sm font-semibold text-foreground">Leírás</h2>
              {/* `whitespace-pre-line`: az admin által írt sortörések
                  megmaradnak, de HTML/markdown nem renderelődik. */}
              <p className="mt-2 whitespace-pre-line text-sm leading-relaxed text-muted-foreground">
                {algorithm.description}
              </p>
            </section>
          )}
        </div>

        {video && (
          <section aria-label="Videó">
            <VideoPanel video={video} title={algorithm.name} />
          </section>
        )}
      </div>

      <AlgorithmPager
        siblings={siblings}
        currentId={algorithm.id}
        caseName={learnCase.name}
      />
    </>
  );
}
