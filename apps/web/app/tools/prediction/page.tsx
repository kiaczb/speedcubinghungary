import type { Metadata } from "next";
import PageHeader from "@/app/components/ui/PageHeader";
import PredictionList from "./components/PredictionList";
import { serverApiUrl } from "@/app/lib/api";
import { PredictionContest } from "./types";

export const metadata: Metadata = {
  title: "Tippverseny",
  description:
    "Tippelj a közelgő versenyek eredményeire, és mérd össze a megérzéseidet a többiekével.",
};

async function getPredictions(): Promise<PredictionContest[]> {
  try {
    const res = await fetch(`${serverApiUrl()}/predictions`, {
      cache: "no-store",
    });
    if (!res.ok) return [];
    return res.json();
  } catch (error) {
    console.error("Nem sikerült elérni a backendet:", error);
    return [];
  }
}

export default async function PredictionPage() {
  const predictions = await getPredictions();

  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-5xl">
        <PageHeader
          title="Tippverseny"
          description="Tippelj a közelgő versenyek eredményeire, és mérd össze a megérzéseidet a többiekével."
        />

        <div className="mt-10">
          <PredictionList predictions={predictions} />
        </div>
      </div>
    </div>
  );
}
