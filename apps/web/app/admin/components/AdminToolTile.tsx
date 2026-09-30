import Link from "next/link";
import Card from "@/app/components/Cards/Card";
import Badge from "@/app/components/Badge";
import type { AdminTool } from "../types";

type Props = {
  tool: AdminTool;
};

export default function AdminToolTile({ tool }: Props) {
  const { title, description, href, icon: Icon, status } = tool;
  const isActive = status === "active";

  const body = (
    <Card
      interactive={isActive}
      padding="lg"
      className={`flex h-full flex-col gap-4 ${isActive ? "" : "opacity-70"}`}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex h-11 w-11 items-center justify-center rounded-control bg-surface-muted text-foreground">
          <Icon className="h-5 w-5" />
        </div>
        {!isActive && (
          <Badge
            label="Hamarosan"
            color="bg-surface-muted text-muted-foreground ring-line"
          />
        )}
      </div>

      <div>
        <h3 className="font-semibold text-foreground">{title}</h3>
        <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
          {description}
        </p>
      </div>
    </Card>
  );

  if (!isActive) {
    return (
      <div aria-disabled className="cursor-not-allowed select-none">
        {body}
      </div>
    );
  }

  return (
    <Link
      href={href}
      className="block rounded-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground"
    >
      {body}
    </Link>
  );
}
