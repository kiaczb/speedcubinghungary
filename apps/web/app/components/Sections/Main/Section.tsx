import React from "react";

type SectionProps = {
  eyebrow?: string;
  title?: string;
  subtitle?: string;
  left?: boolean;
  /** Saját konténert visz a tartalom (teljes szélességű háttérhez). */
  bleed?: boolean;
  children: React.ReactNode;
  className?: string;
};

const Section = ({
  eyebrow,
  title,
  subtitle,
  left = false,
  bleed = false,
  children,
  className = "",
}: SectionProps) => {
  const align = left ? "text-left" : "text-center mx-auto";

  return (
    <section className={`${bleed ? "" : "px-5 sm:px-8 lg:px-12"} ${className}`}>
      <div className={bleed ? "" : "mx-auto w-full max-w-6xl"}>
        {(eyebrow || title || subtitle) && (
          <div className={`mb-10 max-w-2xl ${align}`}>
            {eyebrow && (
              <p className="mb-3 font-mono text-xs font-medium uppercase tracking-[0.18em] text-subtle-foreground">
                {eyebrow}
              </p>
            )}
            {title && (
              <h2 className="text-3xl font-bold text-foreground sm:text-4xl">
                {title}
              </h2>
            )}
            {subtitle && (
              <p className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
                {subtitle}
              </p>
            )}
          </div>
        )}
        {children}
      </div>
    </section>
  );
};

export default Section;
