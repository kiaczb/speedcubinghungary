import { Competition } from "@/app/types";
import { FaCalendarAlt, FaMapMarkerAlt } from "react-icons/fa";
import Card from "../Cards/Card";
import RemainingDays from "./RemainingDays";

// A `Date.now()` render közben tisztátalan hívás lenne (React Compiler szabály),
// ezért a nap-számítás itt, az adatbetöltésben történik.
async function getNextCompetition(): Promise<{
  next: Competition | null;
  daysUntil: number | null;
}> {
  const now = new Date();
  const today = now.toISOString().split("T")[0];

  try {
    const res = await fetch(
      `https://www.worldcubeassociation.org/api/v0/competitions?sort=start_date&start=${today}&country_iso2=HU`,
    );
    if (!res.ok) return { next: null, daysUntil: null };

    const competitions: Competition[] = await res.json();
    const next = competitions?.[0] ?? null;

    return {
      next,
      daysUntil: next
        ? Math.ceil(
            (new Date(next.start_date).getTime() - now.getTime()) /
              (1000 * 60 * 60 * 24),
          )
        : null,
    };
  } catch (error) {
    console.error("Közelgő versenyek lekérése sikertelen:", error);
    return { next: null, daysUntil: null };
  }
}

const UpcomingEvent = async () => {
  const { next, daysUntil } = await getNextCompetition();

  return (
    <Card
      variant="inverse"
      padding="lg"
      className="w-full max-w-sm shrink-0 sm:max-w-md"
    >
      <div className="text-center">
        <h2 className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-white/40">
          Következő verseny
        </h2>
        <div className="mx-auto mt-3 h-px w-12 bg-white/15" />

        <p className="mt-5 text-2xl font-bold leading-tight text-white sm:text-3xl">
          {next ? next.name : "Nincs meghirdetett verseny"}
        </p>

        {next ? (
          <div className="mt-5 flex flex-col items-center gap-2 text-sm text-white/70">
            <span className="flex items-center gap-2">
              <FaCalendarAlt className="h-3.5 w-3.5 shrink-0 text-white/40" />
              <span className="tabular">{next.date_range}</span>
            </span>
            <span className="flex items-center gap-2">
              <FaMapMarkerAlt className="h-3.5 w-3.5 shrink-0 text-white/40" />
              <span>
                {[next.city, next.country_iso2].filter(Boolean).join(", ")}
              </span>
            </span>
          </div>
        ) : (
          <p className="mt-5 text-sm leading-relaxed text-white/60">
            Amint kiírunk egy új versenyt, itt jelenik meg elsőként.
          </p>
        )}
      </div>

      {daysUntil !== null && <RemainingDays days={daysUntil} />}
    </Card>
  );
};

export default UpcomingEvent;
