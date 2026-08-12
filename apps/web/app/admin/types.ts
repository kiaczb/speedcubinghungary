import type { IconType } from "react-icons";

export type AdminToolStatus = "active" | "coming-soon";

export type AdminTool = {
  id: string;
  title: string;
  description: string;
  href: string;
  icon: IconType;
  status: AdminToolStatus;
};
