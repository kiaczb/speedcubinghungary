import { FaChartBar } from "react-icons/fa";
import type { AdminTool } from "../types";

export const ADMIN_TOOLS: AdminTool[] = [
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
