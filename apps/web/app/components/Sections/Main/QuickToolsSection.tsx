// components/Sections/Main/QuickToolsSection.jsx
import Link from "next/link";
import {
  FaListAlt,
  FaUserFriends,
  FaTrophy,
  FaChartLine,
  FaQuestionCircle,
  FaCalendarAlt,
  FaChevronRight,
} from "react-icons/fa";
import Section from "./Section";

const QuickToolsSection = () => {
  const tools = [
    {
      title: "Versenynaptár",
      description: "Tervezett és közelgő versenyek egy helyen",
      meta: "Naptár",
      icon: <FaCalendarAlt className="text-lg" />,
      href: "/calendar",
    },
    {
      title: "Eredmények",
      description: "Korábbi versenyek hivatalos eredményei",
      meta: "Archívum",
      icon: <FaTrophy className="text-lg" />,
      href: "/results",
    },
    {
      title: "Ranglista",
      description: "Országos és regionális rangsorok",
      meta: "Statisztika",
      icon: <FaChartLine className="text-lg" />,
      href: "/rankings",
    },
    {
      title: "Útmutató",
      description: "Lépésről lépésre az első versenyedig",
      meta: "Kezdőknek",
      icon: <FaQuestionCircle className="text-lg" />,
      href: "/guide",
    },
    {
      title: "Csapatunk",
      description: "Ismerd meg a szervezőcsapatot",
      meta: "Egyesület",
      icon: <FaUserFriends className="text-lg" />,
      href: "/team",
    },
    {
      title: "Jelentkezés",
      description: "Nevezés a következő versenyre",
      meta: "Regisztráció",
      icon: <FaListAlt className="text-lg" />,
      href: "/registration",
    },
  ];

  return (
    <Section
      title="Eszközök"
      // subtitle="Gyors hozzáférés a legfontosabb funkciókhoz egy helyen."
      left={true}
    >
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {tools.map((tool, index) => (
          <Link key={index} href={tool.href} className="block h-full">
            <div className="flex h-full flex-col rounded-2xl border border-gray-200 bg-white p-5 transition-shadow duration-200 hover:shadow-md">
              <div className="flex items-center justify-between">
                <div className="flex h-11 w-11 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 text-gray-600">
                  {tool.icon}
                </div>
                <span className="text-xs font-medium uppercase tracking-wide text-gray-400">
                  {tool.meta}
                </span>
              </div>

              <div className="mt-4">
                <h3 className="font-semibold text-gray-900">{tool.title}</h3>
                <p className="mt-1 text-sm leading-relaxed text-gray-500">
                  {tool.description}
                </p>
              </div>

              <div className="mt-4 flex items-center gap-1.5 text-sm font-medium text-gray-700">
                <span>Megnyitás</span>
                <FaChevronRight className="h-3 w-3 text-gray-400" />
              </div>
            </div>
          </Link>
        ))}
      </div>
    </Section>
  );
};

export default QuickToolsSection;
