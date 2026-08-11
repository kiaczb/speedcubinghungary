import Hero from "./components/Hero";
import AboutSection from "./components/Sections/Main/AboutSection";
import CompetitionsSection from "./components/Sections/Main/CompetitionsSection";
import NewsSection from "./components/Sections/Main/NewsSection";
import QuickToolsSection from "./components/Sections/Main/QuickToolsSection";

export default function Home() {
  return (
    <div className="bg-canvas">
      <Hero />

      <div className="flex flex-col gap-20 py-20 sm:gap-24 sm:py-24">
        <AboutSection />
        <CompetitionsSection />
      </div>

      <NewsSection />

      <div className="py-20 sm:py-24">
        <QuickToolsSection />
      </div>
    </div>
  );
}
