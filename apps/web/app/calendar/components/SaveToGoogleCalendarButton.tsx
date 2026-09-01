import { FaRegCalendarPlus } from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import { buildGoogleCalendarUrl } from "../lib/googleCalendar";
import Button from "@/app/components/ui/Button";

type Props = {
  competition: UpcomingCompetition;
};

const SaveToGoogleCalendarButton = ({ competition }: Props) => {
  return (
    <Button
      href={buildGoogleCalendarUrl(competition)}
      external
      variant="secondary"
      size="sm"
      icon={<FaRegCalendarPlus className="h-3.5 w-3.5 text-subtle-foreground" />}
    >
      Mentés Google Naptárba
    </Button>
  );
};

export default SaveToGoogleCalendarButton;
