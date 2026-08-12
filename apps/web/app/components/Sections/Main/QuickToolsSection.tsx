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
import Card from "../../Cards/Card";

const TOOLS = [
  {
    title: "Versenynaptár",
    description: "Tervezett és közelgő versenyek egy helyen",
    meta: "Naptár",
    icon: FaCalendarAlt,
    href: "/calendar",
  },
  {
    title: "Eredmények",
    description: "Korábbi versenyek hivatalos eredményei",
    meta: "Archívum",
    icon: FaTrophy,
    href: "/results",
  },
  {
    title: "Ranglista",
    description: "Országos és regionális rangsorok",
    meta: "Statisztika",
    icon: FaChartLine,
    href: "/rankings",
  },
  {
    title: "Útmutató",
    description: "Lépésről lépésre az első versenyedig",
    meta: "Kezdőknek",
    icon: FaQuestionCircle,
    href: "/guide",
  },
  {
    title: "Csapatunk",
    description: "Ismerd meg a szervezőcsapatot",
    meta: "Egyesület",
    icon: FaUserFriends,
    href: "/team",
  },
  {
    title: "Jelentkezés",
    description: "Nevezés a következő versenyre",
    meta: "Regisztráció",
    icon: FaListAlt,
    href: "/registration",
  },
];

const QuickToolsSection = () => {
  return (
    <Section
      eyebrow="Gyors elérés"
      title="Eszközök"
      subtitle="A leggyakrabban használt oldalak egy helyen."
      left={true}
    >
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {TOOLS.map(({ icon: Icon, ...tool }) => (
          <Link
            key={tool.href}
            href={tool.href}
            className="group rounded-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
          >
            <Card interactive padding="lg" className="h-full">
              <div className="flex items-center justify-between">
                <div className="flex h-11 w-11 items-center justify-center rounded-control border border-line bg-surface-muted text-muted-foreground transition-colors duration-200 ease-smooth group-hover:border-foreground group-hover:bg-foreground group-hover:text-background">
                  <Icon className="h-4 w-4" />
                </div>
                <span className="font-mono text-xs font-medium uppercase tracking-[0.14em] text-subtle-foreground">
                  {tool.meta}
                </span>
              </div>

              <h3 className="mt-5 font-semibold text-foreground">
                {tool.title}
              </h3>
              <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                {tool.description}
              </p>

              <span className="mt-4 inline-flex items-center gap-1.5 text-sm font-medium text-foreground">
                Megnyitás
                <FaChevronRight className="h-2.5 w-2.5 text-subtle-foreground transition-transform duration-200 ease-smooth group-hover:translate-x-0.5" />
              </span>
            </Card>
          </Link>
        ))}
      </div>
    </Section>
  );
};

export default QuickToolsSection;
