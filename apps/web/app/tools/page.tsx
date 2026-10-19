import type { Metadata } from "next";
import PageHeader from "../components/ui/PageHeader";
import UnofficialSection from "./components/sections/UnofficialSection";
import OtherSection from "./components/sections/OtherSection";

export const metadata: Metadata = {
  title: "Programok",
  description:
    "Hazai versenysorozatok saját pontszámítással, valamint közösségi játékok a versenyek köré.",
};

export default function Tools() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-6xl">
        <PageHeader
          title="Programok"
          description="A WCA-versenyeken túl: hazai versenysorozatok és különleges versenyformák saját pontszámítással, valamint közösségi játékok, amelyekkel a versenyeket nézőként is követheted."
        />

        <div className="mt-12 flex flex-col gap-14">
          <UnofficialSection />
          <OtherSection />
        </div>
      </div>
    </div>
  );
}
