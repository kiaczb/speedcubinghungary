import { Suspense } from "react";
import type { Metadata } from "next";
import AdminGate from "../../components/AdminGate";
import GenerateView from "../components/GenerateView";

export const metadata: Metadata = {
  title: "Oklevelek készítése",
  description: "Oklevelek készítése egy sablonnal egy versenyhez.",
  robots: { index: false, follow: false },
};

export default function GenerateCertificatesPage() {
  return (
    <div className="px-5 py-10 sm:px-8 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <AdminGate>
          <Suspense fallback={null}>
            <GenerateView />
          </Suspense>
        </AdminGate>
      </div>
    </div>
  );
}
