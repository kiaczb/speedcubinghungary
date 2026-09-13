import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { FaLayerGroup } from "react-icons/fa";
import PageHeader from "@/app/components/ui/PageHeader";
import StateMessage from "@/app/components/ui/StateMessage";
import CaseCard from "../../components/CaseCard";
import LearnBreadcrumbs from "../../components/LearnBreadcrumbs";
import { getLearnEvent } from "../../lib/data";

type PageProps = {
  params: Promise<{ eventId: string }>;
};

export async function generateMetadata({
  params,
}: PageProps): Promise<Metadata> {
  const { eventId } = await params;
  const detail = await getLearnEvent(eventId);

  if (!detail) return { title: "Algoritmusok" };

  return {
    title: detail.event.name,
    description:
      detail.event.description ||
      `${detail.event.name} — tanulható esetek és algoritmusok.`,
  };
}

export default async function LearnEventPage({ params }: PageProps) {
  const { eventId } = await params;
  const detail = await getLearnEvent(eventId);

  if (!detail) {
    notFound();
  }

  const { event, cases } = detail;

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <LearnBreadcrumbs
          className="mb-6"
          items={[
            { label: "Algoritmusok", href: "/learn" },
            { label: event.name },
          ]}
        />

        <PageHeader title={event.name} description={event.description} />

        <div className="mt-10">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-subtle-foreground">
            Esetek
          </h2>

          <div className="mt-4">
            {cases.length === 0 ? (
              <StateMessage
                icon={<FaLayerGroup className="h-5 w-5" />}
                title="Ehhez a versenyszámhoz még nincs eset"
                description="Amint felkerül az első eset, itt fogod látni."
              />
            ) : (
              <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
                {cases.map((learnCase) => (
                  <CaseCard key={learnCase.id} learnCase={learnCase} />
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
