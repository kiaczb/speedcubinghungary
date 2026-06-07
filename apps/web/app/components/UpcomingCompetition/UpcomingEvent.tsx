import { Competition } from "@/app/types";
import Card from "../Cards/Card";
import { FaCalendar, FaCalendarAlt } from "react-icons/fa";
import { RiTimeFill } from "react-icons/ri";
import RemainingDays from "./RemainingDays";

async function getPosts(): Promise<Competition[]> {
  const res = await fetch(
    `https://www.worldcubeassociation.org/api/v0/competitions?sort=start_date&start=${new Date().toISOString().split('T')[0]}&country_iso2=HU`,
  );
  if (!res.ok) throw new Error("Hírek lekérése sikertelen");
  return res.json();
}

const UpcomingEvent = async () => {
  var competitions = await getPosts();

  if (!competitions || competitions.length === 0) {
    return (
      <div className="relative w-full max-w-sm sm:max-w-md mb-20 lg:my-0 lg:mr-10">
        <Card className="bg-gray-800/90 text-white mx-auto lg:mx-0 p-6">
          <div className="text-center space-y-3">
            {/* Főcím - szürke, kisebb, nagybetűs */}
            <div>
              <h3 className="text-sm md:text-md font-semibold text-gray-400 uppercase tracking-widest mb-1">
                Következő verseny
              </h3>
              <div className="w-16 h-0.5 bg-gray-600 mx-auto"></div>
            </div>

            {/* Verseny neve - fehér, nagy, félkövér */}
            <h4 className="text-2xl md:text-3xl lg:text-4xl font-bold text-white leading-tight">
              Nincs Közelgő verseny
            </h4>

            {/* Dátum és helyszín - tagolt megjelenítés */}
          </div>
        </Card>
      </div>
    );
  }
  // amíg nem töltődik az adat

  // Most már van adat, hozzáférhetsz például az első versenyhez:
  const nextCompetition = competitions[0];

  return (
    <div className="relative w-full max-w-sm sm:max-w-md mb-20 lg:my-0 lg:mr-10">
      <Card className="bg-gray-800/90 text-white mx-auto lg:mx-0 p-6">
        <div className="text-center space-y-3">
          {/* Főcím - szürke, kisebb, nagybetűs */}
          <div>
            <h3 className="text-sm md:text-md font-semibold text-gray-400 uppercase tracking-widest mb-1">
              Következő verseny
            </h3>
            <div className="w-16 h-0.5 bg-gray-600 mx-auto"></div>
          </div>

          {/* Verseny neve - fehér, nagy, félkövér */}
          <h4 className="text-2xl md:text-3xl lg:text-4xl font-bold text-white leading-tight">
            {nextCompetition["name"]}
          </h4>

          {/* Dátum és helyszín - tagolt megjelenítés */}
          <div className="pt-2 space-y-2">
            <div className="flex items-center justify-center gap-2 text-gray-300">
              <FaCalendarAlt className="text-gray-400" />
              <span className="font-medium">{nextCompetition.date_range}</span>
            </div>
            <div className="flex items-center justify-center gap-2 text-gray-300">
              <span className="text-gray-400">📍</span>
              <span>{`${nextCompetition.city}, ${nextCompetition.country_iso2}`}</span>
            </div>
          </div>
        </div>
        <RemainingDays startDate={nextCompetition.start_date} />
      </Card>
    </div>
  );
};

export default UpcomingEvent;
