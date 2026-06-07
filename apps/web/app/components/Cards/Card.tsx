import Image from "next/image";
import React from "react";

type CardProps = {
  imageUrl?: string;
  imageClass?: string;
  children: React.ReactNode;
  className?: string; // extra osztályok átadása
};

const Card = ({ children, imageUrl, imageClass, className }: CardProps) => {
  return (
    <div
      className={`rounded-2xl shadow-xl text-center w-full ${className || ""}`}
    >
      {imageUrl && (
        <div className={`flex justify-center mb-2 ${imageClass || ""}`}>
          <Image
            src={imageUrl}
            alt="image"
            width={150}
            height={100}
            className=" rounded-t-lg"
          />
        </div>
      )}

      <div className="flex flex-col justify-between flex-1">{children}</div>
    </div>
  );
};

export default Card;
