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
    <div className="px-5 py-10 sm:px-8 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader
          title="Adminisztráció"
          description="Az oldal kezeléséhez használt eszközök."
        />

        <div className="mt-12">
          <AdminGate />
        </div>
      </div>
    </div>
  );
}
