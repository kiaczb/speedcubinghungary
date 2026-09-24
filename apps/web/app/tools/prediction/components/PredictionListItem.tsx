import Link from "next/link";
import "@cubing/icons";
import {
  FaMapMarkerAlt,
  FaQuestionCircle,
  FaRegCalendarAlt,
  FaUsers,
} from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import Badge from "@/app/components/Badge";
import { eventName } from "@/app/lib/events";
import { PredictionContest, statusBadgeMap } from "../types";
import DeletePredictionButton from "./DeletePredictionButton";

type Props = {
  prediction: PredictionContest;
  isAdmin?: boolean;
};

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("hu-HU", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

const PredictionListItem = ({ prediction, isAdmin }: Props) => {
  const badge = statusBadgeMap[prediction.status];

  return (
    <Card interactive padding="md" className="group relative">
      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5">
          <Badge color={badge.color} label={badge.label} />
          <span className="flex items-center gap-1.5 text-xs text-subtle-foreground">
            <FaUsers className="h-3 w-3" />
            <span className="tabular font-mono">
              {prediction.participantCount} tipp
            </span>
          </span>

          {isAdmin && (
            <DeletePredictionButton
              predictionId={prediction.id}
              competitionName={prediction.competitionName}
              className="ml-auto"
            />
          )}
        </div>

        <h2 className="text-lg font-semibold leading-snug text-foreground">
          <Link
            href={`/tools/prediction/${prediction.id}`}
            className="after:absolute after:inset-0 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-foreground"
          >
            {prediction.competitionName}
          </Link>
        </h2>

        <div className="flex flex-col gap-1.5 text-sm text-muted-foreground sm:flex-row sm:items-center sm:gap-4">
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

        <div className="flex flex-wrap items-center gap-2">
          {prediction.podiumEventIds.map((eventId) => (
            <span
              key={eventId}
              title={eventName(eventId)}
              className="flex items-center gap-1.5 rounded-control border border-line bg-surface-muted px-2 py-1 text-xs text-muted-foreground"
            >
              <span
                className={`cubing-icon event-${eventId} text-[13px] leading-none text-foreground`}
                aria-hidden
              />
              {eventName(eventId)}
            </span>
          ))}

          {prediction.extraQuestions.length > 0 && (
            <span className="flex items-center gap-1.5 rounded-control border border-line bg-surface-muted px-2 py-1 text-xs text-muted-foreground">
              <FaQuestionCircle className="h-3 w-3 text-subtle-foreground" />
              {prediction.extraQuestions.length} extra kérdés
            </span>
          )}
        </div>
      </div>
    </Card>
  );
};

export default PredictionListItem;
