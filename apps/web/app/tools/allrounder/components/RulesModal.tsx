"use client";

import { useEffect, useState } from "react";
import { FaRegQuestionCircle } from "react-icons/fa";
import { useBodyScrollLock } from "@/app/lib/useBodyScrollLock";
import Button from "@/app/components/ui/Button";

type RulesModalProps = {
  pointsTable: number[];
  countingResults: number;
  podiumCount: number;
};

/** A pontozási szabályok és a helyezésenkénti pontok táblázata. */
const RulesModal = ({
  pointsTable,
  countingResults,
  podiumCount,
}: RulesModalProps) => {
  const [open, setOpen] = useState(false);

  useBodyScrollLock(open);

  useEffect(() => {
    if (!open) return;

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open]);

  // Két oszlopban fut a lista, hogy a 20 helyezés görgetés nélkül elférjen.
  const half = Math.ceil(pointsTable.length / 2);
  const columns = [pointsTable.slice(0, half), pointsTable.slice(half)];

  return (
    <>
      <Button
        variant="secondary"
        size="sm"
        onClick={() => setOpen(true)}
        icon={<FaRegQuestionCircle className="h-4 w-4" />}
      >
        Szabályok
      </Button>

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm dark:bg-black/70"
          onClick={() => setOpen(false)}
          role="presentation"
        >
          <div
            className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-panel border border-line bg-surface shadow-panel"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
            aria-modal="true"
            aria-labelledby="allrounder-rules-title"
          >
            <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-line bg-surface/95 px-6 py-4 backdrop-blur-sm">
              <h2 id="allrounder-rules-title" className="text-lg font-bold">
                Szabályok
              </h2>
              <button
                type="button"
                onClick={() => setOpen(false)}
                aria-label="Bezárás"
                className="-mr-2 -mt-1 inline-flex h-8 w-8 items-center justify-center rounded-control text-xl leading-none text-muted-foreground transition-colors duration-200 ease-smooth hover:bg-surface-muted hover:text-foreground"
              >
                ×
              </button>
            </div>

            <div className="px-6 py-5">
              <ul className="ml-5 list-disc space-y-2 text-sm leading-relaxed text-muted-foreground">
                <li>
                  Minden versenyszám első{" "}
                  <span className="font-semibold text-foreground">
                    {pointsTable.length}
                  </span>{" "}
                  helyezettje kap pontokat.
                </li>
                <li>
                  A sorozat végén a versenyzők legjobb{" "}
                  <span className="font-semibold text-foreground">
                    {countingResults}
                  </span>{" "}
                  versenyszámának pontjait összesítve kapjuk meg a végleges
                  rangsort.
                </li>
                <li>
                  A sorozat végén a pontverseny első{" "}
                  <span className="font-semibold text-foreground">
                    {podiumCount}
                  </span>{" "}
                  helyezettjét oklevéllel és Cuboss kuponkóddal díjazzuk.
                </li>
              </ul>

              <div className="mt-5 grid grid-cols-2 gap-x-6 border-t border-line pt-4">
                {columns.map((column, columnIndex) => (
                  <dl key={columnIndex} className="text-sm">
                    {column.map((points, index) => {
                      const position = columnIndex * half + index + 1;

                      return (
                        <div
                          key={position}
                          className="flex items-baseline justify-between border-b border-line py-1.5 last:border-b-0"
                        >
                          <dt className="tabular font-mono text-muted-foreground">
                            {position}. hely
                          </dt>
                          <dd className="tabular font-mono font-semibold text-foreground">
                            {points} pont
                          </dd>
                        </div>
                      );
                    })}
                  </dl>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default RulesModal;
