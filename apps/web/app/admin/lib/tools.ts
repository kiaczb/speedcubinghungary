import {
  FaChartBar,
  FaGraduationCap,
  FaRegNewspaper,
  FaSatelliteDish,
  FaTrophy,
  FaUsers,
} from "react-icons/fa";
import type { AdminTool } from "../types";

// Minden admin eszköz itt kerül regisztrálásra — új csempe hozzáadásához
// elég ide felvenni egy bejegyzést, a rács automatikusan megjeleníti.
export const ADMIN_TOOLS: AdminTool[] = [
  {
    id: "posts",
    title: "Hírek kezelése",
    description:
      "Posztok, versenybeszámolók és heti riportok létrehozása, szerkesztése és törlése.",
    href: "/posts",
    icon: FaRegNewspaper,
    status: "active",
  },
  {
    id: "special",
    title: "Budapest Special",
    description:
      "A DNF-ek tényleges hosszának rögzítése a Special versenyeken. Az eredménytábla publikus, a versenyt bárki kiválaszthatja.",
    href: "/tools/special",
    icon: FaSatelliteDish,
    status: "active",
  },
  {
    id: "allrounder",
    title: "Hungarian All-Rounder",
    description:
      "A versenysorozat pontversenye. A táblázat publikus; adminként az évad eredményeit újra lehet tölteni a WCA API-ról.",
    href: "/tools/allrounder",
    icon: FaTrophy,
    status: "active",
  },
  {
    id: "learn",
    title: "Algoritmusok",
    description:
      "A tanulófelület tartalma: versenyszámok, esetek és algoritmusok kezelése, opcionális videóval.",
    href: "/admin/learn",
    icon: FaGraduationCap,
    status: "active",
  },
  {
    id: "users",
    title: "Felhasználók",
    description:
      "Regisztrált felhasználók áttekintése és szerepkörök kezelése.",
    href: "/admin/users",
    icon: FaUsers,
    status: "coming-soon",
  },
  {
    id: "stats",
    title: "Statisztikák",
    description:
      "Versenyzői rekordok és versenyzési adatok áttekintése egy helyen.",
    href: "/admin/stats",
    icon: FaChartBar,
    status: "coming-soon",
  },
];
