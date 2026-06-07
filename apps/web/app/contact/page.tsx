import React from "react";

import { FaEnvelope, FaFacebook, FaGlobe } from "react-icons/fa";
import ContactForm from "./components/ContactForm";

const Contact = () => {
  return (
    <div className="p-5 lg:p-12">
      <div className="flex justify-center">
        <h1 className="text-4xl font-bold mb-10">Kapcsolat</h1>
      </div>
      <div className="flex flex-col md:flex-row md:justify-around">
        <div className="w-full p-5">
          <h2 className="font-bold text-2xl mb-3">Speedcubing Hungary</h2>
          <p className="mb-4 max-w-lg">
            Kérdéseid vannak a kockázással kapcsolatban? Szeretnél versenyt
            szervezni vagy részt venni egy eseményen? Vedd fel velünk a
            kapcsolatot e-mailben vagy a honlapunkon!
          </p>

          <ul className="space-y-2 text-lg">
            <li className="flex items-center gap-3 whitespace-nowrap">
              <FaEnvelope className="w-5 h-5 shrink-0" />
              <span>info@speedcubinghungary.hu</span>
            </li>

            <li className="flex items-center gap-3 whitespace-nowrap">
              <FaGlobe className="w-5 h-5 shrink-0" />
              <span>www.speedcubinghungary.hu</span>
            </li>

            <li className="flex items-center gap-3 whitespace-nowrap">
              <FaFacebook className="w-5 h-5 shrink-0" />
              <span>SpeedcubingHungary</span>
            </li>
          </ul>
        </div>
        <div className="w-full p-5 ">
          <ContactForm />
        </div>
      </div>
    </div>
  );
};

export default Contact;
