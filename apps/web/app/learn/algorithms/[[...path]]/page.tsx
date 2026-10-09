import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { serverJson } from "@/app/lib/api";
import LevelView from "../../components/LevelView";
import { ALGORITHMS_DESCRIPTION, levelApiPath } from "../../lib/nodes";
import type { LearnLevel } from "../../types";

type PageProps = {
  params: Promise<{
    path?: string[];
  }>;
};

const EMPTY_ROOT: LearnLevel = {
  node: null,
  ancestors: [],
  children: [],
  algorithms: [],
};

async function getLevel(path: string[]): Promise<LearnLevel | null> {
  return serverJson<LearnLevel | null>(
    levelApiPath(path),
    path.length === 0 ? EMPTY_ROOT : null,
  );
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { path = [] } = await params;
  const level = await getLevel(path);

  if (!level?.node) {
    return { title: "Algoritmusok", description: ALGORITHMS_DESCRIPTION };
  }

  const trail = [...level.ancestors, level.node].map((n) => n.title);
  return {
    title: [...trail].reverse().join(" · "),
    description: level.node.description || `${trail.join(" › ")} — algoritmusok`,
  };
}

export default async function LearnLevelPage({ params }: PageProps) {
  const { path = [] } = await params;
  const level = await getLevel(path);
  if (!level) {
    notFound();
  }

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <LevelView level={level} />
      </div>
    </div>
  );
}
