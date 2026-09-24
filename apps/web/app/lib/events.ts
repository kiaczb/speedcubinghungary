/**
 * WCA versenyszám-azonosító → magyar név.
 *
 * ⚠️ Ez KÖZÖS, nem toolonkénti. Korábban az allrounder és a tippverseny is
 * vitte a saját, szó szerint egyező másolatát — ez nem az egyik tool
 * terméklogikája, hanem a WCA szókincse. Két példánynál egy új versenyszám
 * felvétele az egyik helyen megtörténik, a másikon pedig némán a nyers
 * azonosító jelenik meg a felületen ("sq1" a "Square-1" helyett).
 *
 * A sorrend a WCA saját versenyszám-sorrendje.
 */
const EVENT_NAMES: Record<string, string> = {
  "222": "2x2x2",
  "333": "3x3x3",
  "444": "4x4x4",
  "555": "5x5x5",
  "666": "6x6x6",
  "777": "7x7x7",
  "333bf": "3x3x3 vakon",
  "333fm": "3x3x3 legkevesebb lépés",
  "333oh": "3x3x3 egykezes",
  clock: "Clock",
  minx: "Megaminx",
  pyram: "Pyraminx",
  skewb: "Skewb",
  sq1: "Square-1",
  "444bf": "4x4x4 vakon",
  "555bf": "5x5x5 vakon",
  "333mbf": "3x3x3 többkockás vakon",
};

/** Ismeretlen azonosítóra önmagát adja vissza — a WCA bővítheti a listát. */
export function eventName(eventId: string): string {
  return EVENT_NAMES[eventId] ?? eventId;
}
