import type { CertificateMode } from "../types";
import type { MODES } from "../lib/recipients";

export default function ModeTabs({
  modes,
  mode,
  onChange,
}: {
  modes: typeof MODES;
  mode: CertificateMode;
  onChange: (mode: CertificateMode) => void;
}) {
  const active = modes.find((m) => m.mode === mode);

  return (
    <div className="flex flex-col gap-3">
      {modes.length > 1 && (
        <div
          role="group"
          aria-label="Kik kapnak oklevelet?"
          className="flex rounded-control border border-line bg-surface p-1 sm:inline-flex sm:self-start"
        >
          {modes.map((m) => (
            <button
              key={m.mode}
              type="button"
              aria-pressed={m.mode === mode}
              onClick={() => onChange(m.mode)}
              className={`flex-1 rounded-lg px-4 py-1.5 text-sm font-medium transition-colors duration-200 ease-smooth sm:flex-none ${
                m.mode === mode
                  ? "bg-foreground text-background shadow-card"
                  : "text-muted-foreground hover:bg-surface-muted hover:text-foreground"
              }`}
            >
              {m.label}
            </button>
          ))}
        </div>
      )}
      {active && <p className="text-sm text-muted-foreground">{active.description}</p>}
    </div>
  );
}
