import type { Metadata } from "next";
import CalendarView from "./components/CalendarView";
import PageHeader from "../components/ui/PageHeader";

export const metadata: Metadata = {
  title: "Versenynaptár",
  description:
    "Közelgő magyar WCA versenyek egy helyen, Google Naptár mentéssel.",
};

export default function CalendarPage() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-4xl">
        <PageHeader
          title="Versenynaptár"
          description="Közelgő magyar WCA versenyek időrendben. Mentsd el a számodra érdekes versenyeket egyenesen a Google Naptáradba."
        />

        <div className="mt-10">
          <CalendarView />
        </div>
      </div>
    </div>
  );
}
