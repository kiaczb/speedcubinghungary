// components/Sections/Main/EventsSection.jsx
import Link from "next/link";
import { FaCalendarAlt, FaMapMarkerAlt, FaClock } from "react-icons/fa";
import Section from "./Section";

const EventsSection = () => {
  const upcomingEvents = [
    {
      id: 1,
      title: "Budapest Open 2024",
      date: "2024. november 15-17.",
      location: "Budapest, BME Q épület",
      type: "WCA hivatalos",
      status: "registration_open",
    },
    {
      id: 2,
      title: "Debrecen Winter 2024",
      date: "2024. december 7.",
      location: "Debrecen, Kölcsey Központ",
      type: "WCA hivatalos",
      status: "announced",
    },
  ];

  return (
    <Section title="Közelgő versenyek">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {upcomingEvents.map((event) => (
          <div
            key={event.id}
            className="bg-white rounded-2xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-300 border-l-4 border-green-500"
          >
            <div className="flex justify-between items-start mb-4">
              <h3 className="text-xl font-bold">{event.title}</h3>
              <span className="bg-green-100 text-green-800 text-xs font-semibold px-3 py-1 rounded-full">
                {event.status === "registration_open"
                  ? "Jelentkezés nyitva"
                  : "Hamarosan"}
              </span>
            </div>

            <div className="space-y-3 mb-6">
              <div className="flex items-center text-gray-700">
                <FaCalendarAlt className="mr-3 text-green-600" />
                <span>{event.date}</span>
              </div>
              <div className="flex items-center text-gray-700">
                <FaMapMarkerAlt className="mr-3 text-red-600" />
                <span>{event.location}</span>
              </div>
              <div className="flex items-center text-gray-700">
                <FaClock className="mr-3 text-blue-600" />
                <span>{event.type}</span>
              </div>
            </div>

            <div className="flex gap-4">
              <Link
                href={`/events/${event.id}`}
                className="btn-primary bg-green-600 text-white font-semibold py-2 px-4 rounded-lg text-sm transition duration-300 hover:bg-green-700 flex-1 text-center"
              >
                Részletek
              </Link>
              {event.status === "registration_open" && (
                <Link
                  href={`/registration/${event.id}`}
                  className="border-2 border-green-600 text-green-600 font-semibold py-2 px-4 rounded-lg text-sm transition duration-300 hover:bg-green-50 flex-1 text-center"
                >
                  Jelentkezés
                </Link>
              )}
            </div>
          </div>
        ))}
      </div>

      <div className="text-center mt-8">
        <Link
          href="/events"
          className="inline-flex items-center text-green-700 font-semibold hover:text-green-800 transition duration-300"
        >
          Összes verseny megtekintése
          <svg
            className="w-5 h-5 ml-2"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M17 8l4 4m0 0l-4 4m4-4H3"
            />
          </svg>
        </Link>
      </div>
    </Section>
  );
};

export default EventsSection;
