import type { Metadata } from "next";
import { FaArrowRight } from "react-icons/fa";
import PageHeader from "../components/ui/PageHeader";
import Card from "../components/Cards/Card";
import Button from "../components/ui/Button";

export const metadata: Metadata = {
  title: "Rólunk",
  description:
    "A Speedcubing Hungary Egyesület 2018 óta szervez hivatalos WCA versenyeket Magyarországon.",
};

const STATS = [
  { value: "2018", label: "Alapítás éve" },
  { value: "50+", label: "Megszervezett verseny" },
  { value: "5000+", label: "Résztvevő összesen" },
  { value: "15", label: "Város adott otthont" },
];

export default function About() {
  return (
    <div className="px-5 py-12 sm:px-8 sm:py-16 lg:px-12">
      <div className="mx-auto w-full max-w-3xl">
        <PageHeader
          title="Rólunk"
          description="Kik vagyunk, mit csinálunk, és hogyan csatlakozhatsz."
        />

        <div className="mt-10 space-y-5 text-lg leading-relaxed text-muted-foreground">
          <p>
            A{" "}
            <strong className="font-semibold text-foreground">
              Speedcubing Hungary Egyesület
            </strong>{" "}
            2018 óta működik, és az ország legnagyobb kockázó közösségét
            képviseli. WCA regionális szervezetként hivatalos versenyeket
            szervezünk, ahol a kezdőktől a világbajnokokig mindenki megtalálja a
            számára megfelelő kihívásokat.
          </p>
          <p>
            Célunk, hogy Magyarországon népszerűsítsük a speedcubing sportot,
            biztosítsuk a versenyzők számára a legjobb körülményeket, és
            építsünk egy erős, támogató közösséget.
          </p>
          <p>
            Versenyeink nyitottak: nem kell gyorsnak lenned ahhoz, hogy elindulj.
            A legtöbben az első versenyükön ismerik meg a közösséget — és utána
            maradnak.
          </p>
        </div>

        <div className="mt-10 grid grid-cols-2 gap-4 sm:grid-cols-4">
          {STATS.map((stat) => (
            <Card key={stat.label} padding="md">
              <p className="tabular font-mono text-2xl font-bold leading-none text-foreground">
                {stat.value}
              </p>
              <p className="mt-2 text-xs leading-relaxed text-muted-foreground">
                {stat.label}
              </p>
            </Card>
          ))}
        </div>

        <div className="mt-10 flex flex-wrap gap-3 border-t border-line pt-8">
          <Button
            href="/calendar"
            iconRight={<FaArrowRight className="h-3.5 w-3.5" />}
          >
            Közelgő versenyek
          </Button>
          <Button href="/contact" variant="secondary">
            Írj nekünk
          </Button>
        </div>
      </div>
    </div>
  );
}
