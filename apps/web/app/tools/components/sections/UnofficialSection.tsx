import React from "react";
import { FaGlobe } from "react-icons/fa";
import ToolCard from "../ToolCard";

const UnofficialSection = () => {
  return (
    <section>
      <h2 className="text-xl font-bold text-foreground">
        Versenysorozatok és versenyformák
      </h2>
      <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground">
        Saját pontszámítású sorozatok és formátumok, a WCA hivatalos
        helyezésein túl.
      </p>

      <div className="mt-6 grid grid-cols-1 items-stretch gap-5 md:grid-cols-2">
        <ToolCard
          title="Budapest Special"
          links={[
            {
              href: "/tools/special",
              label: "Megnyitás",
              icon: <FaGlobe className="h-4 w-4" />,
            },
          ]}
        >
          <p>
            Itt érhetők el a Special végződésű versenyek eredményei. A
            versenyeken az összes versenyszám idejének összege számít
            versenyzőnként. Az élő verseny eredményei a WCA Live-ról érkeznek.
          </p>
        </ToolCard>

        <ToolCard
          title="Hungarian All-Rounder"
          links={[
            {
              href: "/tools/allrounder",
              label: "Megnyitás",
              icon: <FaGlobe className="h-4 w-4" />,
            },
          ]}
        >
          <p>
            A Hungarian All-Rounder egy állomásokból álló versenysorozat, melyen
            a WCA-helyezéseken felül a versenyzők összesített eredményeit is
            díjazzuk. Az egyes állomások külön-külön WCA-versenyekként
            funkcionálnak, ezért mindig külön kell rájuk regisztrálni.
          </p>
        </ToolCard>
      </div>
    </section>
  );
};

export default UnofficialSection;
