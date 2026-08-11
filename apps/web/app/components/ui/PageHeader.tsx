import React from "react";

type PageHeaderProps = {
  eyebrow?: string;
  title: string;
  description?: string;
  className?: string;
};

const PageHeader = ({
  eyebrow,
  title,
  description,
  className = "",
}: PageHeaderProps) => {
  return (
    <header className={`border-b border-line pb-8 ${className}`}>
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
    </header>
  );
};

export default PageHeader;
