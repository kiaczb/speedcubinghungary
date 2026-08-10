import React from "react";

/**
 * Űrlapmező-keret: címke + vezérlő + súgószöveg.
 *
 * A kapcsolat-űrlap és a poszt-modal külön-külön stílusozta az input-jait (más
 * kerekítés, más fókuszgyűrű), a kapcsolat-űrlap címkéi ráadásul nem is voltak
 * a mezőkhöz kötve — kattintásra és képernyőolvasóval sem működtek.
 * A `htmlFor` itt kötelező, így ez nem fordulhat elő újra.
 */

/** Közös input/textarea/select osztályok. Add át a vezérlő `className`-jének. */
export const controlClass =
  "w-full rounded-control border border-line bg-surface px-3.5 py-2.5 text-sm text-foreground " +
  "placeholder:text-subtle-foreground " +
  "transition-[border-color,box-shadow] duration-200 ease-smooth " +
  "focus:border-foreground focus:outline-none focus:ring-2 focus:ring-foreground/10";

type FieldProps = {
  label: string;
  /** A vezérlő `id`-ja — ezzel köti össze a címkét és a mezőt. */
  htmlFor: string;
  /** Kiegészítő magyarázat a mező alatt. */
  hint?: React.ReactNode;
  /** Halvány "(opcionális)" jelölés a címke mellett. */
  optional?: boolean;
  children: React.ReactNode;
  className?: string;
};

const Field = ({
  label,
  htmlFor,
  hint,
  optional,
  children,
  className = "",
}: FieldProps) => {
  return (
    <div className={className}>
      <label
        htmlFor={htmlFor}
        className="mb-1.5 block text-sm font-medium text-foreground"
      >
        {label}
        {optional && (
          <span className="ml-1.5 font-normal text-subtle-foreground">
            (opcionális)
          </span>
        )}
      </label>
      {children}
      {hint && (
        <p className="mt-1.5 text-xs leading-relaxed text-muted-foreground">
          {hint}
        </p>
      )}
    </div>
  );
};

export default Field;
