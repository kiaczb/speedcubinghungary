import type { Metadata } from "next";
import PageHeader from "@/app/components/ui/PageHeader";
import AdminGate from "../components/AdminGate";
import LearnAdminView from "./components/LearnAdminView";

export const metadata: Metadata = {
  title: "Algoritmusok kezelése",
  description: "A tanulófelület tartalmának kezelése.",
  robots: { index: false, follow: false },
};

export default function AdminLearnPage() {
  return (
    <div className="px-5 py-10 sm:px-8 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader
          title="Algoritmusok kezelése"
          description="A tanulófelület tartalma: versenyszámok, azokon belül esetek, az eseteken belül algoritmusok — opcionális videóval. Az olvasás mindenki számára nyitott, a szerkesztés csak adminoknak."
        />

        <div className="mt-12">
          {/* Ugyanaz a kliensoldali RBAC-kapu, mint az admin főoldalon. A
              tényleges védelmet a backend adja (JWT + RBAC minden írásnál). */}
          <AdminGate>
            <LearnAdminView />
          </AdminGate>
        </div>
      </div>
    </div>
  );
}
