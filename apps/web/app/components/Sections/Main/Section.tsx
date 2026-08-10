import React from "react";

/**
 * A főoldal szakaszainak közös kerete.
 *
 * Változások a korábbi verzióhoz képest:
 * - a `className` hiányakor a class-listába szó szerint bekerült az "undefined";
 * - minden szakasz maga döntötte el a belső szélességét és a padding-jét
 *   (`px-10 lg:px-30` itt, `container mx-auto px-4` ott), így a szakaszok
 *   tartalma nem volt egy vonalban. Most közös konténer tartja a rendet.
 */

type SectionProps = {
  /** Rövid, nagybetűs felirat a cím fölött. */
  eyebrow?: string;
  title?: string;
  subtitle?: string;
  /** `left` megtartva a régi hívásokhoz: balra igazított fejléc. */
  left?: boolean;
  /**
   * Saját konténert visz a tartalom (pl. teljes szélességű háttér esetén),
   * ezért a szakasz nem tesz rá vízszintes paddinget.
   */
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
