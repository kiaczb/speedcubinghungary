import type { Metadata } from "next";
import { FaGraduationCap } from "react-icons/fa";
import PageHeader from "@/app/components/ui/PageHeader";
import StateMessage from "@/app/components/ui/StateMessage";
import EventCard from "../components/EventCard";
import { getLearnEvents } from "../lib/data";

export const metadata: Metadata = {
  title: "Algoritmusok",
  description:
    "Tanulható algoritmusok versenyszámok és esetek szerint rendezve, videós magyarázatokkal.",
};

export default async function LearnPage() {
  const events = await getLearnEvents();

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <PageHeader
          title="Algoritmusok"
          description="Válassz versenyszámot, azon belül esetet, és tanuld meg a hozzá tartozó algoritmusokat. Ahol van hozzá videó, ott az algoritmus mellett meg is nézheted."
        />

        <div className="mt-10">
          {events.length === 0 ? (
            <StateMessage
              icon={<FaGraduationCap className="h-5 w-5" />}
              title="Még nincs feltöltött anyag"
              description="Amint felkerül az első versenyszám, itt fogod látni."
            />
          ) : (
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
              {events.map((event) => (
                <EventCard key={event.id} event={event} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
