// A lista szerveroldalon kér adatot (`no-store`), ezért navigáláskor van egy
// rövid várakozás — ezt tölti ki ez a vázlat.
//
// ⚠️ Ezért van az `(overview)` útvonalcsoport: egy `loading.tsx` a `learn`
// szegmensben az ALATTA lévő oldalakra is érvényes lenne, azok pedig a
// streamelés miatt 200-zal adnák vissza a „nincs ilyen” oldalt a 404 helyett
// (a státusz már elment, mire a `notFound()` lefut). A csoport miatt a vázlat
// csak a listára vonatkozik, a részletoldalak pedig valódi 404-et adnak.
// Ne told fel ezt a fájlt egy szinttel.
export default function LearnLoading() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div
        role="status"
        aria-label="Betöltés…"
        className="mx-auto w-full max-w-5xl animate-pulse"
      >
        <div className="h-4 w-48 rounded-control bg-surface-muted" />

        <div className="mt-8 border-b border-line pb-8">
          <div className="h-9 w-2/3 rounded-control bg-surface-muted sm:w-1/2" />
          <div className="mt-4 h-4 w-full rounded-control bg-surface-muted" />
          <div className="mt-2 h-4 w-4/5 rounded-control bg-surface-muted" />
        </div>

        <div className="mt-10 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              key={index}
              className="h-32 rounded-card border border-line bg-surface"
            />
          ))}
        </div>
      </div>
    </div>
  );
}
