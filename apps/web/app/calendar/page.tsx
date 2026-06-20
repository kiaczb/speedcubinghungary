import type { Metadata } from "next";
import CalendarList from "./components/CalendarList";

export const metadata: Metadata = {
  title: "Versenynaptár | Speedcubing Hungary",
  description: "Közelgő magyar WCA versenyek egy helyen, Google Naptár mentéssel.",
};

export default function CalendarPage() {
  return (
    <div className="min-h-screen bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl">
        <header className="mb-10">
          <h1 className="text-3xl font-bold text-gray-900 sm:text-4xl">
            Versenynaptár
          </h1>
          <p className="mt-3 text-gray-600">
            Közelgő magyar WCA versenyek időrendben. Mentsd el a számodra
            érdekes versenyeket egyenesen a Google Naptáradba.
          </p>
        </header>

        <CalendarList />
      </div>
    </div>
  );
}
