import { Suspense } from "react";
import type { Metadata } from "next";
import PageHeader from "@/app/components/ui/PageHeader";
import SpecialView from "./components/SpecialView";

export const metadata: Metadata = {
  title: "Budapest Special",
  description:
    "A Special versenyek élő eredménytáblája: minden versenyszám idejének összege versenyzőnként, a DNF-ek tényleges hosszával együtt.",
};

export default function SpecialPage() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader
          title="Budapest Special"
          description="A Special versenyeken az összes versenyszám idejének összege számít. Válaszd ki a versenyt, és az eredmények élőben érkeznek a WCA Live-ról — a DNF-ek tényleges hosszát a helyszínen rögzítjük, mert azt a hivatalos eredmény nem őrzi meg."
        />

        <div className="mt-10">
          <Suspense fallback={null}>
            <SpecialView />
          </Suspense>
        </div>
      </div>
    </div>
  );
}
