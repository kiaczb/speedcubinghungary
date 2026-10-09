import { CubingIcons } from "@cubing/icons/js";
import { eventName } from "@/app/lib/events";

const EVENT_PREFIX = "event-";
const UNOFFICIAL_PREFIX = "unofficial-";

const ALL_ICONS: string[] = Object.values(CubingIcons);

export const EVENT_ICONS = ALL_ICONS.filter((id) => id.startsWith(EVENT_PREFIX));
export const UNOFFICIAL_ICONS = ALL_ICONS.filter((id) =>
  id.startsWith(UNOFFICIAL_PREFIX),
);

export function iconLabel(icon: string): string {
  if (icon.startsWith(EVENT_PREFIX)) return eventName(icon.slice(EVENT_PREFIX.length));
  return icon.slice(UNOFFICIAL_PREFIX.length).replaceAll("_", " ");
}
