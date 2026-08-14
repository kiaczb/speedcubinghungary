import { Competition } from "@/app/types";
import { FaCalendarAlt, FaMapMarkerAlt } from "react-icons/fa";

async function getNextCompetition(): Promise<Competition | null> {
  const today = new Date().toISOString().split("T")[0];

  try {
    const res = await fetch(
      `https://www.worldcubeassociation.org/api/v0/competitions?sort=start_date&start=${today}&country_iso2=HU`,
    );
    if (!res.ok) return null;

    const competitions: Competition[] = await res.json();
    return competitions?.[0] ?? null;
  } catch (error) {
    console.error("Közelgő versenyek lekérése sikertelen:", error);
    return null;
  }
}

const UpcomingEvent = async () => {
  const next = await getNextCompetition();

  return (
    <div className="w-full max-w-sm shrink-0 rounded-card border border-line-strong bg-surface p-6 shadow-card sm:max-w-md sm:p-7">
      <p className="font-mono text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
        Következő verseny
      </p>

      <div className="mt-4">
        {next ? (
          <>
            <p className="text-2xl font-bold leading-snug text-foreground sm:text-3xl">
              {next.name}
            </p>
            <div className="mt-3 flex flex-col gap-1.5 text-sm text-muted-foreground">
              <span className="flex items-center gap-2">
                <FaCalendarAlt className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                <span className="tabular">{next.date_range}</span>
              </span>
              <span className="flex items-center gap-2">
                <FaMapMarkerAlt className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                <span>
                  {[next.city, next.country_iso2].filter(Boolean).join(", ")}
                </span>
              </span>
            </div>
          </>
        ) : (
          <p className="text-sm leading-relaxed text-muted-foreground">
            Nincs meghirdetett verseny. Amint kiírunk egyet, itt jelenik meg
            elsőként.
          </p>
        )}
      </div>
    </div>
  );
};

export default UpcomingEvent;
