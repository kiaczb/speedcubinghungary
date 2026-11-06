import { formatCentiseconds } from "../lib/time";
import type { RoundLimit } from "../lib/limits";

type RoundLimitsProps = {
  limits: RoundLimit[];
};

const RoundLimits = ({ limits }: RoundLimitsProps) => {
  if (limits.length === 0) return null;

  return (
    <section className="flex flex-wrap items-baseline gap-x-8 gap-y-1 rounded-card border border-line bg-surface px-4 py-3">
      {limits.map((limit) => (
        <p key={limit.key} className="flex items-baseline gap-2">
          <span className="text-sm text-muted-foreground">
            {limits.length > 1 ? `${limit.label} limitje` : "Limit"}
          </span>
          <span className="tabular font-mono text-lg font-semibold text-foreground">
            {formatCentiseconds(limit.limitCs)}
          </span>
        </p>
      ))}

      <p className="text-sm text-muted-foreground">
        az összes kategóriára együtt
      </p>
    </section>
  );
};

export default RoundLimits;
