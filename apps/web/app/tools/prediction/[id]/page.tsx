import Link from "next/link";
import { notFound } from "next/navigation";
import {
  FaArrowLeft,
  FaMapMarkerAlt,
  FaRegCalendarAlt,
  FaUsers,
} from "react-icons/fa";
import Badge from "@/app/components/Badge";
import Button from "@/app/components/ui/Button";
import { serverApiUrl } from "@/app/lib/api";
import PodiumEventForm from "../components/PodiumEventForm";
import ExtraQuestionField from "../components/ExtraQuestionField";
import { PredictionContest, statusBadgeMap } from "../types";

type PageProps = {
  params: Promise<{
    id: string;
  }>;
};

async function getPrediction(id: string): Promise<PredictionContest | null> {
  try {
    // Nincs még GET /predictions/{id} végpont — a listából szűrünk, amíg a
    // backend csak a listázást és a létrehozást támogatja (lásd CLAUDE.md §21).
    const res = await fetch(`${serverApiUrl()}/predictions`, {
      cache: "no-store",
    });
    if (!res.ok) return null;

    const predictions: PredictionContest[] = await res.json();
    return predictions.find((p) => p.id === id) ?? null;
  } catch (error) {
    console.error("Nem sikerült elérni a backendet:", error);
    return null;
  }
}

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("hu-HU", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

function formatDeadline(date: string): string {
  return new Date(date).toLocaleString("hu-HU", {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default async function PredictionDetailPage({ params }: PageProps) {
  const { id } = await params;
  const prediction = await getPrediction(id);
  if (!prediction) {
    notFound();
  }

  const badge = statusBadgeMap[prediction.status];
  const isOpen = prediction.status === "open";

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-3xl">
        <Link
          href="/tools/prediction"
          className="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
        >
          <FaArrowLeft className="h-3 w-3 transition-transform duration-200 ease-smooth group-hover:-translate-x-0.5" />
          Vissza a tippversenyekhez
        </Link>

        <header className="mt-6 border-b border-line pb-8">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
            <Badge color={badge.color} label={badge.label} />
            <span className="flex items-center gap-1.5 text-xs text-subtle-foreground">
              <FaUsers className="h-3 w-3" />
              <span className="tabular font-mono">
                {prediction.participantCount} tipp
              </span>
            </span>
          </div>

          <h1 className="mt-4 text-3xl font-bold leading-tight text-foreground sm:text-4xl">
            {prediction.competitionName}
          </h1>

          <div className="mt-3 flex flex-col gap-1.5 text-sm text-muted-foreground sm:flex-row sm:items-center sm:gap-4">
            <span className="flex items-center gap-2">
              <FaRegCalendarAlt className="h-3.5 w-3.5 shrink-0 text-subtle-foreground" />
              <span className="tabular">
                {formatDate(prediction.competitionStartDate)}
              </span>
            </span>
            <span className="flex items-center gap-2">
              <FaMapMarkerAlt className="h-3.5 w-3.5 shrink-0 text-subtle-foreground" />
              {[prediction.competitionCity, prediction.competitionCountry]
                .filter(Boolean)
                .join(", ")}
            </span>
          </div>

          <p className="mt-4 text-sm text-muted-foreground">
            {isOpen ? (
              <>
                Tippelési határidő:{" "}
                <span className="tabular font-medium text-foreground">
                  {formatDeadline(prediction.predictionDeadline)}
                </span>
              </>
            ) : (
              <>
                A tippelés lezárult{" "}
                <span className="tabular font-medium text-foreground">
                  {formatDeadline(prediction.predictionDeadline)}
                </span>
                -kor.
              </>
            )}
          </p>
        </header>

        <form className="mt-8 flex flex-col gap-10">
          <section>
            <h2 className="text-lg font-semibold text-foreground">
              Dobogó-tippek
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Add meg, szerinted ki végez az 1-2-3. helyen az egyes
              versenyszámokban.
            </p>

            <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
              {prediction.podiumEventIds.map((eventId) => (
                <PodiumEventForm key={eventId} eventId={eventId} />
              ))}
            </div>
          </section>

          {prediction.extraQuestions.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-foreground">
                Extra kérdések
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Ezekre a kérdésekre is tippelhetsz.
              </p>

              <div className="mt-4 flex flex-col gap-4">
                {prediction.extraQuestions.map((question) => (
                  <ExtraQuestionField key={question.id} question={question} />
                ))}
              </div>
            </section>
          )}

          <div className="flex justify-end border-t border-line pt-6">
            <Button type="submit" disabled={!isOpen}>
              {isOpen ? "Tipp leadása" : "A tippelés lezárult"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
