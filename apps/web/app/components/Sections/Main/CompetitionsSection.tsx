import Image from "next/image";
import React from "react";
import Section from "./Section";
import { FaArrowRight, FaGraduationCap } from "react-icons/fa";
import Link from "next/link";

const CompetitionsSection = () => {
  return (
    <Section title="Versenyzés" left={false}>
      <div className="flex flex-wrap justify-around gap-10 mt-20">
        <div>
          <Image
            alt="dog"
            src="/wanna_start_cubing.jpg"
            width={650}
            height={650}
          />
        </div>
        <div className="flex flex-col justify-center text-xl gap-10 max-w-md">
          <p>
            Érdekel a versenyzés világa? Ki tudod rakni a kockát és szeretnéd
            megpróbáltattni magad egy versenyen? Ne habozz és látogass el az
            általunk rendezett versenyek egyikére.
          </p>
          <p>
            Esetleg attól félsz, hogy nem vagy elég gyors? Ez ne tántorítson el,
            a kockások egy rendkíül befogadó közösség! Azonban ha fejlsődni
            szeretnél oldalunkon megtalálhatóak oktató anyagok ahol az ország
            legjobb versenyzőitől tanulhatsz!
          </p>
        </div>
        <div className="flex flex-col gap-10 sm:gap-0 sm:flex-row justify-around w-full items-center">
          <Link
            href="/competitions"
            className="group inline-flex items-center px-6 py-3 bg-black text-white font-semibold rounded-lg border border-gray-800 hover:border-gray-600 transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5 duration-300"
          >
            <span>Versenyezni szeretnék</span>
            <FaArrowRight className="w-4 h-4 ml-3 transform group-hover:translate-x-1 transition-transform duration-300" />
          </Link>
          <Link
            href="/competitions"
            className="group inline-flex items-center px-6 py-3  text-black font-semibold rounded-lg border border-gray-800 hover:border-gray-600 transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5 duration-300"
          >
            <FaGraduationCap className="w-6 h-6 mr-5" />

            <span>Tanulni szeretnék</span>
            <FaArrowRight className="w-4 h-4 ml-3 transform group-hover:translate-x-1 transition-transform duration-300" />
          </Link>
        </div>
      </div>
    </Section>
  );
};

export default CompetitionsSection;
