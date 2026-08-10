import type { Metadata } from "next";
import PageHeader from "../components/ui/PageHeader";
import UnofficialSection from "./components/sections/UnofficialSection";
import OtherSection from "./components/sections/OtherSection";

export const metadata: Metadata = {
  title: "Eszközök",
  description:
    "Versenyekhez és versenysorozatokhoz kapcsolódó alkalmazások, alternatív rangsorok és statisztikák.",
};

export default function Tools() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <PageHeader
          eyebrow="Alkalmazások"
          title="Eszközök"
          description="Itt találod azokat az alkalmazásokat és aloldalakat, amelyek versenyekhez és versenysorozatokhoz kapcsolódnak: eredmények feldolgozása, összesítése és alternatív rangsorok. Egy részük speciális pontszámításra vagy összetett statisztikákra épül, amelyek a WCA hivatalos helyezésein túlmutatva adnak részletesebb képet a versenyzők teljesítményéről."
        />

        <div className="mt-12 flex flex-col gap-14">
          <UnofficialSection />
          <OtherSection />
        </div>
      </div>
    </div>
  );
}
