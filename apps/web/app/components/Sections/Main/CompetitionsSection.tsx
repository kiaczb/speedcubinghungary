import Image from "next/image";
import React from "react";
import Section from "./Section";
import Button from "../../ui/Button";
import { FaArrowRight, FaGraduationCap } from "react-icons/fa";

const CompetitionsSection = () => {
  return (
    <Section eyebrow="Kezdés" title="Versenyzés" left={true}>
      <div className="grid grid-cols-1 items-center gap-12 lg:grid-cols-2 lg:gap-16">
        <div className="overflow-hidden rounded-panel border border-line">
          <Image
            alt="Kockázók egy magyar versenyen"
            src="/wanna_start_cubing.jpg"
            width={650}
            height={650}
            className="h-full w-full object-cover"
          />
        </div>

        <div>
          <p className="text-lg leading-relaxed text-muted-foreground">
            Érdekel a versenyzés világa? Ki tudod rakni a kockát és szeretnéd
            megpróbáltatni magad egy versenyen? Ne habozz, látogass el az
            általunk rendezett versenyek egyikére.
          </p>
          <p className="mt-5 text-lg leading-relaxed text-muted-foreground">
            Attól félsz, hogy nem vagy elég gyors? Ez ne tántorítson el — a
            kockások rendkívül befogadó közösség. Ha pedig fejlődnél, oldalunkon
            oktatóanyagokat is találsz, ahol az ország legjobb versenyzőitől
            tanulhatsz.
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Button
              href="/calendar"
              iconRight={<FaArrowRight className="h-3.5 w-3.5" />}
            >
              Versenyezni szeretnék
            </Button>
            <Button
              href="/tools"
              variant="secondary"
              icon={<FaGraduationCap className="h-4 w-4" />}
              iconRight={<FaArrowRight className="h-3.5 w-3.5" />}
            >
              Tanulni szeretnék
            </Button>
          </div>
        </div>
      </div>
    </Section>
  );
};

export default CompetitionsSection;
