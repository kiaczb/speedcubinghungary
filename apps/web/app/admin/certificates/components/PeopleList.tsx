import { countryName } from "../../stats/lib/format";
import type { Recipient } from "../types";

export default function PeopleList({
  recipients,
  isExcluded,
  onToggle,
}: {
  recipients: Recipient[];
  isExcluded: (recipient: Recipient) => boolean;
  onToggle: (recipient: Recipient) => void;
}) {
  return (
    <ul className="flex flex-col divide-y divide-line rounded-card border border-line bg-surface">
      {recipients.map((recipient, index) => {
        const included = !isExcluded(recipient);
        return (
          <li key={`${recipient.wcaId || recipient.name}-${index}`}>
            <label
              className={`flex cursor-pointer items-center gap-4 px-4 py-2.5 transition-colors duration-200 ease-smooth hover:bg-surface-muted ${
                included ? "" : "opacity-50"
              }`}
            >
              <input
                type="checkbox"
                checked={included}
                onChange={() => onToggle(recipient)}
                className="h-4 w-4 shrink-0 accent-foreground"
              />
              <span className="tabular w-8 shrink-0 text-right font-mono text-xs text-subtle-foreground">
                {index + 1}.
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">{recipient.name}</span>
                <span className="tabular block font-mono text-xs text-muted-foreground">
                  {recipient.wcaId || "újonc"}
                </span>
              </span>
              <span className="hidden shrink-0 text-sm text-muted-foreground sm:block">
                {countryName(recipient.countryIso2)}
              </span>
            </label>
          </li>
        );
      })}
    </ul>
  );
}
