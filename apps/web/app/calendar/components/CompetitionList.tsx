import { FaRegCalendarTimes } from "react-icons/fa";
import { UpcomingCompetition } from "../types";
import CompetitionCard from "./CompetitionCard";
import StateMessage from "./StateMessage";

type Props = {
  competitions: UpcomingCompetition[];
};

const CompetitionList = ({ competitions }: Props) => {
  if (competitions.length === 0) {
    return (
      <StateMessage
        icon={<FaRegCalendarTimes className="h-5 w-5" />}
        title="Nincs közelgő verseny"
        description="Jelenleg nincs meghirdetett magyar verseny. Nézz vissza később!"
      />
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {competitions.map((competition) => (
        <CompetitionCard key={competition.id} competition={competition} />
      ))}
    </div>
  );
};

export default CompetitionList;
