import type { Metadata } from "next";
import AdminGate from "../../../components/AdminGate";
import TemplateLoader from "./components/TemplateLoader";

export const metadata: Metadata = {
  title: "Sablonszerkesztő",
  description: "Oklevélsablon szerkesztése.",
  robots: { index: false, follow: false },
};

export default async function CertificateTemplatePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <AdminGate>
      <TemplateLoader id={id} />
    </AdminGate>
  );
}
