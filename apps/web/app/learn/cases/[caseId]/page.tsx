import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { FaListUl } from "react-icons/fa";
import PageHeader from "@/app/components/ui/PageHeader";
import StateMessage from "@/app/components/ui/StateMessage";
import AlgorithmListItem from "../../components/AlgorithmListItem";
import LearnBreadcrumbs from "../../components/LearnBreadcrumbs";
import { getLearnCase } from "../../lib/data";

type PageProps = {
  params: Promise<{ caseId: string }>;
};

export async function generateMetadata({
  params,
}: PageProps): Promise<Metadata> {
  const { caseId } = await params;
  const detail = await getLearnCase(caseId);

  if (!detail) return { title: "Algoritmusok" };

  return {
    title: `${detail.case.name} — ${detail.event.name}`,
    description:
      detail.case.description ||
      `${detail.case.name} algoritmusai (${detail.event.name}).`,
  };
}

export default async function LearnCasePage({ params }: PageProps) {
  const { caseId } = await params;
  const detail = await getLearnCase(caseId);

  if (!detail) {
    notFound();
  }

  const { event, case: learnCase, algorithms } = detail;

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-4xl">
        <LearnBreadcrumbs
          className="mb-6"
          items={[
            { label: "Algoritmusok", href: "/learn" },
            { label: event.name, href: `/learn/events/${event.id}` },
            { label: learnCase.name },
          ]}
        />

        <PageHeader title={learnCase.name} description={learnCase.description} />

        <div className="mt-10">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-subtle-foreground">
            Algoritmusok
          </h2>

          <div className="mt-4">
            {algorithms.length === 0 ? (
              <StateMessage
                icon={<FaListUl className="h-5 w-5" />}
                title="Ehhez az esethez még nincs algoritmus"
                description="Amint felkerül az első algoritmus, itt fogod látni."
              />
            ) : (
              <div className="flex flex-col gap-4">
                {algorithms.map((algorithm) => (
                  <AlgorithmListItem key={algorithm.id} algorithm={algorithm} />
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
