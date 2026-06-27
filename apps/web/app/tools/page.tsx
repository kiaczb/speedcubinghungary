import UnofficialSection from "./components/sections/UnofficialSection";
import OtherSection from "./components/sections/OtherSection";

export default function Tools() {
  return (
    <div className="px-10 lg:py-20 lg:px-30">
      <div className="mb-10">
        <div className="flex justify-center">
          <h1 className="text-4xl font-bold mb-10">Eszközök</h1>
        </div>
        <p>
          A Tools oldalon találhatóak azok az alkalmazások és aloldalak, amelyek
          különböző versenyekhez és versenysorozatokhoz kapcsolódnak. Ezek az
          eszközök az eredmények feldolgozását, összesítését és alternatív
          rangsorok megjelenítését szolgálják. A toolok egy része speciális
          pontszámításokra vagy összetett statisztikákra épül, amelyek a WCA
          hivatalos helyezésein túlmutatva nyújtanak részletesebb képet a
          versenyzők teljesítményéről.
        </p>
      </div>
      <div>
        <div className="flex flex-col gap-10">
          <UnofficialSection />
          <OtherSection />
        </div>
      </div>
    </div>
  );
}
