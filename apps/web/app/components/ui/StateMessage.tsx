import React from "react";

/**
 * Üres és hibaállapotok egységes megjelenítése.
 *
 * Korábban csak a versenynaptár alatt élt; mivel semmi naptár-specifikus nincs
 * benne, felkerült a közös komponensek közé — így a hírek listája is ugyanezt
 * az üres állapotot kapja, nem egy sajátot.
 */

type StateMessageProps = {
  icon: React.ReactNode;
  title: string;
  description: string;
  /** Kivezető akció — az üres képernyő legyen felszólítás, ne zsákutca. */
  action?: React.ReactNode;
};

const StateMessage = ({
  icon,
  title,
  description,
  action,
}: StateMessageProps) => (
  <div className="flex flex-col items-center justify-center rounded-panel border border-dashed border-line-strong bg-surface px-6 py-16 text-center">
    <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-surface-muted text-subtle-foreground">
      {icon}
    </div>
    <h3 className="text-base font-semibold text-foreground">{title}</h3>
    <p className="mt-1.5 max-w-sm text-sm leading-relaxed text-muted-foreground">
      {description}
    </p>
    {action && <div className="mt-6">{action}</div>}
  </div>
);

export default StateMessage;
