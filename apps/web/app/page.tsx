// app/page.js
import Hero from "./components/Hero";
import AboutSection from "./components/Sections/Main/AboutSection";
import CompetitionsSection from "./components/Sections/Main/CompetitionsSection";
import EventsSection from "./components/Sections/Main/EventsSection";
import NewsSection from "./components/Sections/Main/NewsSection";
import QuickToolsSection from "./components/Sections/Main/QuickToolsSection";

export default function Home() {
  return (
    <div className="flex flex-col bg-gray-50 gap-30">
      <Hero />
      <div className="flex flex-col gap-20">
        <AboutSection />
        <NewsSection />
        <CompetitionsSection />

        {/* <EventsSection /> */}
        <QuickToolsSection />
        {/* Később ide jöhet a ToolsSection */}
      </div>
    </div>
  );
}
