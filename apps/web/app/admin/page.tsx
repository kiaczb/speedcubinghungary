import type { Metadata } from "next";
import PageHeader from "../components/ui/PageHeader";
import AdminGate from "./components/AdminGate";

export const metadata: Metadata = {
  title: "Admin",
  description: "Adminisztrátori eszközök és funkciók.",
  robots: { index: false, follow: false },
};

export default function AdminPage() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <PageHeader
          eyebrow="Admin felület"
          title="Adminisztráció"
          description="Innen érheted el az oldal kezeléséhez használt adminisztrátori eszközöket."
        />

        <div className="mt-12">
          <AdminGate />
        </div>
      </div>
    </div>
  );
}
