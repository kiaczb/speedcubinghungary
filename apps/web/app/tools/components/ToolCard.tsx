import React, { ReactNode } from "react";
import Link from "next/link";
import Card from "@/app/components/Cards/Card";

type ToolLink = {
  href: string;
  label: string;
  icon?: ReactNode;
};

type ToolCardProps = {
  title: string;
  children: React.ReactNode; //description
  links: ToolLink[];
};
const ToolCard = ({ title, children, links }: ToolCardProps) => {
  return (
    <Card className="max-w-lg">
      <div className="flex flex-col items-center mb-6 rounded-t-2xl bg-gray-200 p-2">
        <h1 className="text-xl">{title}</h1>
      </div>

      <div className="flex flex-col gap-14 p-6">
        <div>{children}</div>

        <div className="flex flex-col gap-2 items-center">
          {links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="btn-primary bg-black text-white font-semibold py-3 px-6 rounded-xl flex items-center gap-2 hover:bg-gray-800"
            >
              {link.icon}
              <span className="text-white">{link.label}</span>
            </Link>
          ))}
        </div>
      </div>
    </Card>
  );
};

export default ToolCard;
