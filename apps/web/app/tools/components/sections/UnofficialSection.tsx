import React from "react";
import { FaGlobe } from "react-icons/fa";
import ToolCard from "../ToolCard";

const UnofficialSection = () => {
  return (
    <div>
      <h2 className="text-3xl">Nem hivatalos eredmények</h2>
      <div className="flex flex-wrap justify-center items-start gap-10">
        <ToolCard
          title="Budapest Special"
          links={[
            {
              href: "#events",
              label: "Visit",
              icon: <FaGlobe className="w-5 h-5 shrink-0" />,
            },
          ]}
        >
          <div>
            <p>
              Ezen az oldalon érhetőek el a Special végződésű versenyek
              eredményei. A versenyeken az összes versenyszám idejének összege
              található emberenként
            </p>
          </div>
        </ToolCard>

        <ToolCard
          title="Hungarian All-Rounder"
          links={[
            {
              href: "https://allrounderpoints.vercel.app/",
              label: "Visit",
              icon: <FaGlobe className="w-5 h-5 shrink-0" />,
            },
          ]}
        >
          <div>
            <p>
              A Hungarian All-Rounder egy állomásokból álló versenysorozat,
              melyen (a WCA-helyezéseken felül) a versenyzők összesített
              eredményeit is díjazzuk. Az egyes állomások külön-külön
              WCA-versenyekként funkcionálnak, ezért mindig külön kell rájuk
              regisztrálni.
            </p>
          </div>
        </ToolCard>
      </div>
    </div>
  );
};

export default UnofficialSection;
