// components/Sections/Main/AboutSection.jsx
import Image from "next/image";
import Link from "next/link";
import Section from "./Section";

const AboutSection = () => {
  return (
    <Section title="Rólunk" left={true} className="pb-30">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-10 items-start">
        <div className="text-md">
          <p className="text-gray-700 mb-6 text-lg">
            A <strong>Speedcubing Hungary Egyesület</strong> 2018 óta működik és
            az ország legnagyobb kockázó közösségét képviseli. WCA regionális
            szervezetként hivatalos versenyeket szervezünk, ahol a kezdőktől a
            világbajnokokig mindenki megtalálja a számára megfelelő kihívásokat.
          </p>
          <p className="text-gray-700 mb-8 text-lg">
            Célunk, hogy Magyarországon népszerűsítsük a speedcubing sportot,
            biztosítsuk a versenyzők számára a legjobb körülményeket és építsünk
            egy erős, támogató közösséget.
          </p>

          <div className="flex flex-wrap gap-4">
            <Link
              href="/about"
              className="btn-primary bg-black text-white font-semibold py-3 px-6 rounded-xl transition duration-300 hover:bg-gray-800"
            >
              Bővebben az egyesületről
            </Link>
            <Link
              href="/contact"
              className="border-2 border-black font-semibold py-3 px-6 rounded-xl transition duration-300 hover:bg-gray-100"
            >
              Kapcsolat
            </Link>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="bg-white p-6 rounded-2xl shadow-lg text-center">
            <div className="text-4xl font-bold text-green-600 mb-2">50+</div>
            <div className="text-gray-700">Megszervezett verseny</div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-lg text-center">
            <div className="text-4xl font-bold text-red-600 mb-2">5000+</div>
            <div className="text-gray-700">Résztvevő összesen</div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-lg text-center">
            <div className="text-4xl font-bold text-blue-600 mb-2">10</div>
            <div className="text-gray-700">Aktív szervező</div>
          </div>
          <div className="bg-white p-6 rounded-2xl shadow-lg text-center">
            <div className="text-4xl font-bold text-purple-600 mb-2">15</div>
            <div className="text-gray-700">Városban versenyeztünk</div>
          </div>
        </div>
      </div>
    </Section>
  );
};

export default AboutSection;
