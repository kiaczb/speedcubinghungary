import React from "react";

type SectionProps = {
  title?: string;
  left?: boolean;
  children: React.ReactNode;
  className?: string; // extra osztályok átadása
};

const Section = ({
  title,
  left = false,
  children,
  className,
}: SectionProps) => {
  return (
    <section className={`px-10 lg:px-30 ${className}`}>
      {title && (
        <h1
          className={`text-4xl font-bold mb-6 ${
            left ? "text-left" : "text-center"
          }`}
        >
          {title}
        </h1>
      )}
      {children}
    </section>
  );
};

export default Section;
