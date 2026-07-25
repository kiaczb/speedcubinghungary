import Link from "next/link";
import React from "react";
import { FaEnvelope, FaFacebook, FaGlobe } from "react-icons/fa";

const Footer = () => {
  return (
    <footer className="bg-gray-800 text-white py-6">
      <div className="max-w-6xl mx-auto px-4 grid grid-cols-1 md:grid-cols-3 gap-8 text-center md:text-left">
        <div>
          <h1 className="font-bold text-2xl mb-2">Speedcubing Hungary</h1>
          <p className="text-sm md:text-base">
            A Speedcubing Hungary Egyesület célja a kockázás népszerűsítése és
            WCA regionális szervezetként hivatalos WCA versenyek szervezése.
          </p>
        </div>

        <div>
          <h1 className="font-bold text-xl mb-2">Quick Links</h1>
          <ul className="space-y-1">
            <li>
              <Link href="/tools">Tools</Link>
            </li>
            <li>
              <Link href="/about">About</Link>
            </li>
            <li>
              <Link href="/contact">Contact</Link>
            </li>
            <li>
              <Link href="/posts">Blog</Link>
            </li>
          </ul>
        </div>

        <div>
          <h1 className="font-bold text-xl mb-2">Contact info</h1>
          <ul className="space-y-2">
            <li className="flex items-center justify-center md:justify-start gap-3 whitespace-nowrap">
              <FaEnvelope className="w-5 h-5 shrink-0" />
              <span>info@speedcubinghungary.hu</span>
            </li>

            <li className="flex items-center justify-center md:justify-start gap-3 whitespace-nowrap">
              <FaGlobe className="w-5 h-5 shrink-0" />
              <span>www.speedcubinghungary.hu</span>
            </li>

            <li className="flex items-center justify-center md:justify-start gap-3 whitespace-nowrap">
              <FaFacebook className="w-5 h-5 shrink-0" />
              <span>SpeedcubingHungary</span>
            </li>
          </ul>
        </div>
      </div>

      <div className="mt-6 text-center text-sm">
        <p>
          &copy; {new Date().getFullYear()} Speedcubing Hungary. All rights
          reserved.
        </p>
      </div>
    </footer>
  );
};

export default Footer;
