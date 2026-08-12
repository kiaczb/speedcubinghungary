import { FaChartBar, FaRegNewspaper, FaUsers } from "react-icons/fa";
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
