import Link from "next/link";
import { notFound } from "next/navigation";
import {
  FaArrowLeft,
  FaMapMarkerAlt,
  FaRegCalendarAlt,
  FaUsers,
} from "react-icons/fa";
import Badge from "@/app/components/Badge";
import { serverJson } from "@/app/lib/api";
import { formatDateTime, formatLongDate } from "@/app/lib/date";
import PredictionForm from "../components/PredictionForm";
import { PredictionContest, statusBadgeMap } from "../types";

type PageProps = {
  params: Promise<{
    id: string;
  }>;
};

async function getPrediction(id: string): Promise<PredictionContest | null> {
  const predictions = await serverJson<PredictionContest[]>("/predictions", []);

  return predictions.find((p) => p.id === id) ?? null;
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
                {formatLongDate(prediction.competitionStartDate)}
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
                  {formatDateTime(prediction.predictionDeadline)}
                </span>
              </>
            ) : (
              <>
                A tippelés lezárult{" "}
                <span className="tabular font-medium text-foreground">
                  {formatDateTime(prediction.predictionDeadline)}
                </span>
                -kor.
              </>
            )}
          </p>
        </header>

        <PredictionForm prediction={prediction} isOpen={isOpen} />
      </div>
    </div>
  );
}
