import React, { ReactNode } from "react";
import Card from "@/app/components/Cards/Card";
import Button from "@/app/components/ui/Button";

type ToolLink = {
  href: string;
  label: string;
  icon?: ReactNode;
};

type ToolCardProps = {
  title: string;
  children: React.ReactNode;
  links: ToolLink[];
};

const ToolCard = ({ title, children, links }: ToolCardProps) => {
  return (
    <Card interactive className="flex h-full flex-col">
      <div className="border-b border-line px-5 py-4">
        <h3 className="font-semibold text-foreground">{title}</h3>
      </div>

      <div className="flex flex-1 flex-col gap-6 p-5">
        <div className="flex-1 text-sm leading-relaxed text-muted-foreground">
          {children}
        </div>

        <div className="flex flex-wrap gap-2">
          {links.map((link) => (
            <Button
              key={link.href}
              href={link.href}
              external={link.href.startsWith("http")}
              size="sm"
              icon={link.icon}
            >
              {link.label}
            </Button>
          ))}
        </div>
      </div>
    </Card>
  );
};

export default ToolCard;
