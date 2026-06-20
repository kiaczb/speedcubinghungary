import { FaMapMarkerAlt, FaRegCalendarAlt, FaExternalLinkAlt } from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import SaveToGoogleCalendarButton from "./SaveToGoogleCalendarButton";

type Props = {
  competition: UpcomingCompetition;
};

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("hu-HU", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

function formatDateRange(start: string, end: string): string {
  if (start === end) return formatDate(start);
  return `${formatDate(start)} – ${formatDate(end)}`;
}

const CompetitionCard = ({ competition }: Props) => {
  const startDate = new Date(competition.startDate);
  const month = startDate
    .toLocaleDateString("hu-HU", { month: "short" })
    .replace(".", "")
    .toUpperCase();
  const day = startDate.toLocaleDateString("hu-HU", { day: "numeric" });

  return (
    <div className="flex flex-col gap-5 rounded-2xl border border-gray-200 bg-white p-5 transition-shadow duration-200 hover:shadow-md sm:flex-row sm:gap-6">
      {/* Date chip */}
      <div className="flex h-16 w-16 shrink-0 flex-col items-center justify-center rounded-xl border border-gray-200 bg-gray-50">
        <span className="text-xs font-medium uppercase tracking-wide text-gray-400">
          {month}
        </span>
        <span className="text-2xl font-bold leading-none text-gray-900">
          {day}
        </span>
      </div>

      {/* Details */}
      <div className="flex flex-1 flex-col gap-3">
        <h3 className="text-lg font-semibold text-gray-900">
          {competition.name}
        </h3>

        <div className="flex flex-col gap-1.5 text-sm text-gray-500">
          <span className="flex items-center gap-2">
            <FaRegCalendarAlt className="h-4 w-4 shrink-0 text-gray-400" />
            {formatDateRange(competition.startDate, competition.endDate)}
          </span>
          <span className="flex items-center gap-2">
            <FaMapMarkerAlt className="h-4 w-4 shrink-0 text-gray-400" />
            {[competition.city, competition.country].filter(Boolean).join(", ")}
          </span>
        </div>

        <div className="mt-1 flex flex-wrap items-center gap-3">
          <SaveToGoogleCalendarButton competition={competition} />
          <a
            href={competition.url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-sm font-medium text-gray-700 transition-colors hover:text-gray-900"
          >
            <span>Verseny megnyitása</span>
            <FaExternalLinkAlt className="h-3 w-3 text-gray-400" />
          </a>
        </div>
      </div>
    </div>
  );
};

export default CompetitionCard;
