// components/Sections/Main/QuickToolsSection.jsx
import Link from "next/link";
import {
  FaCalculator,
  FaListAlt,
  FaUserFriends,
  FaTrophy,
  FaChartLine,
  FaQuestionCircle,
  FaCalendarAlt,
} from "react-icons/fa";
import Section from "./Section";

const QuickToolsSection = () => {
  const tools = [
    {
      title: "Versenynaptár",
      description: "Összes tervezett verseny egy helyen",
      icon: <FaCalendarAlt className="text-2xl" />,
      color: "bg-blue-100 text-blue-600",
      href: "/calendar",
    },
    {
      title: "Eredmények",
      description: "Korábbi versenyek eredményei",
      icon: <FaTrophy className="text-2xl" />,
      color: "bg-yellow-100 text-yellow-600",
      href: "/results",
    },
    {
      title: "Ranglista",
      description: "Országos és regionális rangsor",
      icon: <FaChartLine className="text-2xl" />,
      color: "bg-purple-100 text-purple-600",
      href: "/rankings",
    },
    {
      title: "Útmutató",
      description: "Hogyan jelentkezz első versenyedre",
      icon: <FaQuestionCircle className="text-2xl" />,
      color: "bg-green-100 text-green-600",
      href: "/guide",
    },
    {
      title: "Csapatunk",
      description: "Ismerd meg a szervezőket",
      icon: <FaUserFriends className="text-2xl" />,
      color: "bg-red-100 text-red-600",
      href: "/team",
    },
    {
      title: "Jelentkezés",
      description: "Következő versenyre jelentkezés",
      icon: <FaListAlt className="text-2xl" />,
      color: "bg-indigo-100 text-indigo-600",
      href: "/registration",
    },
  ];

  return (
    <Section title="Eszközök" className="mb-20 mt-30" left={true}>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {tools.map((tool, index) => (
          <Link key={index} href={tool.href} className="group">
            <div className="bg-white rounded-2xl shadow-md p-6 hover:shadow-xl transition-all duration-300 hover:-translate-y-1 h-full">
              <div className="flex items-start gap-4">
                <div className={`${tool.color} p-3 rounded-xl`}>
                  {tool.icon}
                </div>
                <div>
                  <h3 className="font-bold text-lg mb-1 group-hover:text-green-700 transition-colors">
                    {tool.title}
                  </h3>
                  <p className="text-gray-600 text-sm">{tool.description}</p>
                </div>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </Section>
  );
};

export default QuickToolsSection;
