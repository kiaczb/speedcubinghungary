import type { Metadata } from "next";
import PageHeader from "@/app/components/ui/PageHeader";
import AdminGate from "../components/AdminGate";
import TemplateGallery from "./components/TemplateGallery";

export const metadata: Metadata = {
  title: "Oklevelek",
  description: "Oklevélsablonok és oklevelek készítése versenyekhez.",
  robots: { index: false, follow: false },
};

export default function AdminCertificatesPage() {
  return (
    <div className="px-5 py-10 sm:px-8 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader
          title="Oklevelek"
          description="Válassz egy sablont, majd egy versenyt, és elkészül a PDF. Új sablont is itt hozhatsz létre."
        />
        <div className="mt-10">
          <AdminGate>
            <TemplateGallery />
          </AdminGate>
        </div>
      </div>
    </div>
  );
}
