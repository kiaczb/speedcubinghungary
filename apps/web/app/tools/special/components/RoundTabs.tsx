"use client";

type RoundTabsProps = {
  /** A választható körsorszámok, növekvő sorrendben. */
  rounds: number[];
  /** A kiválasztott kör, vagy null = összes. */
  selected: number | null;
  onSelect: (round: number | null) => void;
};

const TAB_BASE =
  "rounded-control px-3.5 py-1.5 text-sm font-medium transition-colors duration-200 ease-smooth " +
  "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-foreground";

/**
 * Kör-választó: minden versenyszám adott sorszámú körére szűkít (1 → 333-r1,
 * 222-r1, …). A szűrés a szerveren történik, hogy a felhasznált/maradék idő és
 * a kumulatív keret a meglévő egyetlen implementációval számolódjon.
 *
 * Egyetlen kör esetén nem jelenítünk meg semmit — nincs mit választani.
 */
const RoundTabs = ({ rounds, selected, onSelect }: RoundTabsProps) => {
  if (rounds.length < 2) return null;

  const tabClass = (isActive: boolean) =>
    `${TAB_BASE} ${
      isActive
        ? "bg-foreground text-background"
        : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
    }`;

  return (
    <div
      role="tablist"
      aria-label="Kör kiválasztása"
      className="flex flex-wrap items-center gap-1 rounded-card border border-line bg-surface p-1"
    >
      <button
        type="button"
        role="tab"
        aria-selected={selected === null}
        onClick={() => onSelect(null)}
        className={tabClass(selected === null)}
      >
        Összes
      </button>

      {rounds.map((round) => (
        <button
          key={round}
          type="button"
          role="tab"
          aria-selected={selected === round}
          onClick={() => onSelect(round)}
          className={tabClass(selected === round)}
        >
          {round}. kör
        </button>
      ))}
    </div>
  );
};

export default RoundTabs;
