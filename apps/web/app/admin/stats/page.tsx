import { Suspense } from "react";
import type { Metadata } from "next";
import PageHeader from "@/app/components/ui/PageHeader";
import AdminGate from "../components/AdminGate";
import StatsView from "./components/StatsView";

export const metadata: Metadata = {
  title: "Statisztikák",
  description: "Versenystatisztikák adminisztrátoroknak.",
  robots: { index: false, follow: false },
};

export default function AdminStatsPage() {
  return (
    <div className="px-5 py-10 sm:px-8 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader
          title="Statisztikák"
          description="Egy verseny érdekességei: előtte a nevezők alapján, közben a WCA Live-ról, utána a hivatalos eredményekből."
        />

        <div className="mt-10">
          <AdminGate>
            <Suspense fallback={null}>
              <StatsView />
            </Suspense>
          </AdminGate>
        </div>
      </div>
    </div>
  );
}
