import {
  FaMapMarkerAlt,
  FaRegCalendarAlt,
  FaExternalLinkAlt,
} from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import Card from "@/app/components/Cards/Card";
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
    <Card
      interactive
      padding="md"
      className="flex flex-col gap-5 sm:flex-row sm:gap-6"
    >
      <div className="flex h-16 w-16 shrink-0 flex-col items-center justify-center rounded-control border border-line bg-surface-muted">
        <span className="font-mono text-[10px] font-medium uppercase tracking-[0.14em] text-subtle-foreground">
          {month}
        </span>
        <span className="tabular font-mono text-2xl font-bold leading-none text-foreground">
          {day}
        </span>
      </div>

      <div className="flex flex-1 flex-col gap-3">
        <h3 className="font-semibold text-foreground">{competition.name}</h3>

        <div className="flex flex-col gap-1.5 text-sm text-muted-foreground">
          <span className="flex items-center gap-2">
            <FaRegCalendarAlt className="h-3.5 w-3.5 shrink-0 text-subtle-foreground" />
            <span className="tabular">
              {formatDateRange(competition.startDate, competition.endDate)}
            </span>
          </span>
          <span className="flex items-center gap-2">
            <FaMapMarkerAlt className="h-3.5 w-3.5 shrink-0 text-subtle-foreground" />
            {[competition.city, competition.country].filter(Boolean).join(", ")}
          </span>
        </div>

        <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-3">
          <SaveToGoogleCalendarButton competition={competition} />
          <a
            href={competition.url}
            target="_blank"
            rel="noopener noreferrer"
            className="group inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
          >
            <span>Verseny megnyitása</span>
            <FaExternalLinkAlt className="h-2.5 w-2.5 text-subtle-foreground transition-transform duration-200 ease-smooth group-hover:-translate-y-0.5" />
          </a>
        </div>
      </div>
    </Card>
  );
};

export default CompetitionCard;
