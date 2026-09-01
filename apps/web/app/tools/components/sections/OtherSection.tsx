import React from "react";
import { FaGlobe } from "react-icons/fa";
import ToolCard from "../ToolCard";

const OtherSection = () => {
  return (
    <section>
      <h2 className="text-xl font-bold text-foreground">Egyéb</h2>
      <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground">
        Közösségi eszközök a versenyek köré.
      </p>

      <div className="mt-6 grid grid-cols-1 items-stretch gap-5 md:grid-cols-2">
        {/* TODO: ideiglenes leírások — cseréld le a véglegesre. */}
        <ToolCard
          title="Szavazás"
          links={[
            {
              href: "#events",
              label: "Megnyitás",
              icon: <FaGlobe className="h-4 w-4" />,
            },
          ]}
        >
          <p>
            Közösségi szavazások a versenyekkel és az egyesület döntéseivel
            kapcsolatban.
          </p>
        </ToolCard>

        <ToolCard
          title="Tippverseny"
          links={[
            {
              href: "#events",
              label: "Megnyitás",
              icon: <FaGlobe className="h-4 w-4" />,
            },
          ]}
        >
          <p>
            Tippelj a közelgő versenyek eredményeire, és mérd össze a
            megérzéseidet a többiekével.
          </p>
        </ToolCard>
      </div>
    </section>
  );
};

export default OtherSection;
