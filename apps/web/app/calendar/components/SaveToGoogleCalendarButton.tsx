import { FaRegCalendarPlus } from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import { buildGoogleCalendarUrl } from "../lib/googleCalendar";

type Props = {
  competition: UpcomingCompetition;
};

const SaveToGoogleCalendarButton = ({ competition }: Props) => {
  return (
    <a
      href={buildGoogleCalendarUrl(competition)}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
    >
      <FaRegCalendarPlus className="h-4 w-4 text-gray-500" />
      <span>Mentés Google Naptárba</span>
    </a>
  );
};

export default SaveToGoogleCalendarButton;
