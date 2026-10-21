import {
  FaChartBar,
  FaRegNewspaper,
  FaSatelliteDish,
  FaTrophy,
  FaUsers,
} from "react-icons/fa";
import type { AdminTool } from "../types";

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
      "Egy verseny statisztikái előtte (nevezők), közben (WCA Live) és utána (hivatalos eredmények): visszatérők, évfordulók, mérföldkövek, rekordok.",
    href: "/admin/stats",
    icon: FaChartBar,
    status: "active",
  },
];
