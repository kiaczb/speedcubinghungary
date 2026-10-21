import Hero from "./components/Hero";
import AboutSection from "./components/Sections/Main/AboutSection";
import CompetitionsSection from "./components/Sections/Main/CompetitionsSection";
import NewsSection from "./components/Sections/Main/NewsSection";
import HighlightsSection from "./components/Sections/Main/HighlightsSection";

export default function Home() {
  return (
    <div className="bg-canvas">
      <Hero />

      <div className="pb-20 pt-12 sm:pb-24 sm:pt-16">
        <HighlightsSection />
      </div>

      <NewsSection />

      <div className="flex flex-col gap-20 py-20 sm:gap-24 sm:py-24">
        <AboutSection />
        <CompetitionsSection />
      </div>
    </div>
  );
}
