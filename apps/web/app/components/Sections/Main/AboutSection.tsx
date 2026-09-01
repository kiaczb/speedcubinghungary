import Section from "./Section";
import Card from "../../Cards/Card";
import Button from "../../ui/Button";

const STATS = [
  { value: "50+", label: "Megszervezett verseny", accent: "text-brand-red" },
  { value: "5000+", label: "Résztvevő összesen", accent: "text-brand-green" },
  { value: "10", label: "Aktív szervező", accent: "text-brand-red" },
  { value: "15", label: "Város adott otthont", accent: "text-brand-green" },
];

const AboutSection = () => {
  return (
    <Section eyebrow="Egyesület" title="Rólunk" left={true}>
      <div className="grid grid-cols-1 items-center gap-12 lg:grid-cols-2 lg:gap-16">
        <div>
          <p className="text-lg leading-relaxed text-muted-foreground">
            A <strong className="font-semibold text-foreground">Speedcubing
            Hungary Egyesület</strong> 2018 óta működik és az ország legnagyobb
            kockázó közösségét képviseli. WCA regionális szervezetként hivatalos
            versenyeket szervezünk, ahol a kezdőktől a világbajnokokig mindenki
            megtalálja a számára megfelelő kihívásokat.
          </p>
          <p className="mt-5 text-lg leading-relaxed text-muted-foreground">
            Célunk, hogy Magyarországon népszerűsítsük a speedcubing sportot,
            biztosítsuk a versenyzők számára a legjobb körülményeket és építsünk
            egy erős, támogató közösséget.
          </p>

          <div className="mt-8 flex flex-wrap gap-3">
            <Button href="/about">Bővebben az egyesületről</Button>
            <Button href="/contact" variant="secondary">
              Kapcsolat
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {STATS.map((stat) => (
            <Card key={stat.label} padding="lg">
              <p
                className={`tabular font-mono text-4xl font-bold leading-none ${stat.accent}`}
              >
                {stat.value}
              </p>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                {stat.label}
              </p>
            </Card>
          ))}
        </div>
      </div>
    </Section>
  );
};

export default AboutSection;
