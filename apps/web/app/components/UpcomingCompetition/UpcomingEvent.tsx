import { FaCalendarAlt, FaMapMarkerAlt } from "react-icons/fa";
import { UpcomingCompetition } from "@/app/calendar/types";
import { serverJson } from "@/app/lib/api";
import { formatDateRange } from "@/app/lib/date";

// A saját /competitions/upcoming végpontunkon át, nem közvetlenül a WCA-tól:
// ott van a WCA-hívás egyetlen helye (lapozás, gyorsítótár, párhuzamossági
// korlát), és így ugyanazt mutatjuk, amit a naptár.
async function getNextCompetition(): Promise<UpcomingCompetition | null> {
  const competitions = await serverJson<UpcomingCompetition[]>(
    "/competitions/upcoming",
    [],
  );
  return competitions[0] ?? null;
}

const UpcomingEvent = async () => {
  const next = await getNextCompetition();

  return (
    <div className="w-full max-w-sm shrink-0 rounded-card border border-line-strong bg-surface p-6 shadow-card sm:max-w-md sm:p-7">
      <p className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
        Következő verseny
      </p>

     {next ? (
       <a
         href={next.url}
         target="_blank"
         rel="noopener noreferrer"
       >
         <div className="mt-4">
           <p className="text-2xl font-bold leading-snug text-foreground sm:text-3xl">
             {next.name}
           </p>
           <div className="mt-3 flex flex-col gap-1.5 text-sm text-muted-foreground">
             <span className="flex items-center gap-2">
               <FaCalendarAlt className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
               <span className="tabular">
                 {formatDateRange(next.startDate, next.endDate)}
               </span>
             </span>
             <span className="flex items-center gap-2">
               <FaMapMarkerAlt className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
               <span>
                 {[next.city, next.country].filter(Boolean).join(", ")}
               </span>
             </span>
           </div>
         </div>
       </a>
     ) : (
       <div className="mt-4">
         <p className="text-sm leading-relaxed text-muted-foreground">
           Nincs meghirdetett verseny. Amint kiírunk egyet, itt jelenik meg
           elsőként.
         </p>
       </div>
     )}
    </div>
  );
};

export default UpcomingEvent;
