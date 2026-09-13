import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { FaArrowLeft } from "react-icons/fa";
import AlgorithmStudy from "../../components/AlgorithmStudy";
import LearnBreadcrumbs from "../../components/LearnBreadcrumbs";
import { getLearnAlgorithm } from "../../lib/data";

type PageProps = {
  params: Promise<{ algorithmId: string }>;
};

export async function generateMetadata({
  params,
}: PageProps): Promise<Metadata> {
  const { algorithmId } = await params;
  const detail = await getLearnAlgorithm(algorithmId);

  if (!detail) return { title: "Algoritmusok" };

  return {
    title: `${detail.algorithm.name} — ${detail.case.name}`,
    description:
      detail.algorithm.description ||
      `${detail.algorithm.name}: ${detail.algorithm.notation}`,
  };
}

export default async function LearnAlgorithmPage({ params }: PageProps) {
  const { algorithmId } = await params;
  const detail = await getLearnAlgorithm(algorithmId);

  if (!detail) {
    notFound();
  }

  const { event, case: learnCase, algorithm } = detail;

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <LearnBreadcrumbs
          className="mb-6"
          items={[
            { label: "Algoritmusok", href: "/learn" },
            { label: event.name, href: `/learn/events/${event.id}` },
            { label: learnCase.name, href: `/learn/cases/${learnCase.id}` },
            { label: algorithm.name },
          ]}
        />

        <Link
          href={`/learn/cases/${learnCase.id}`}
          className="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors duration-200 ease-smooth hover:text-foreground"
        >
          <FaArrowLeft className="h-3 w-3 transition-transform duration-200 ease-smooth group-hover:-translate-x-0.5" />
          Vissza az esethez
        </Link>

        <header className="mt-6 border-b border-line pb-8">
          <h1 className="text-3xl font-bold text-foreground sm:text-4xl">
            {algorithm.name}
          </h1>
          <p className="mt-3 text-base text-muted-foreground">
            {event.name} → {learnCase.name}
          </p>
        </header>

        <div className="mt-10">
          <AlgorithmStudy detail={detail} />
        </div>
      </div>
    </div>
  );
}
