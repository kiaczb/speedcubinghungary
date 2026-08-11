import React from "react";

type PageHeaderProps = {
  eyebrow?: string;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
};

const PageHeader = ({
  eyebrow,
  title,
  description,
  action,
  className = "",
}: PageHeaderProps) => {
  return (
    <header
      className={`flex flex-col gap-5 border-b border-line pb-8 sm:flex-row sm:items-end sm:justify-between ${className}`}
    >
      <div className="max-w-2xl">
        {eyebrow && (
          <p className="mb-2 font-mono text-xs font-medium uppercase tracking-[0.18em] text-subtle-foreground">
            {eyebrow}
          </p>
        )}
        <h1 className="text-3xl font-bold text-foreground sm:text-4xl">
          {title}
        </h1>
        {description && (
          <p className="mt-3 text-base leading-relaxed text-muted-foreground">
            {description}
          </p>
        )}
      </div>
      {action && <div className="flex shrink-0 items-center gap-3">{action}</div>}
    </header>
  );
};

export default PageHeader;
