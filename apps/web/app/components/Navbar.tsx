"use client";

import Image from "next/image";
import Link from "next/link";
import React, { useEffect, useState } from "react";
import SpeedcubingFullSVG from "./svgs/SpeedcubingFullSVG";
import SpeedcubingFullAnimatedSVG from "./svgs/SpeedcubingFullAnimatedSVG";

type NavbarProps = { logoUrl?: string; children?: React.ReactNode };

const Navbar = ({ logoUrl, children }: NavbarProps) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 0);
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);
  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  return (
    <nav
      className={`
    sticky top-0 w-full z-50
    bg-white dark:bg-gray-800
    transition-all duration-100 ease-in
    ${isScrolled ? "shadow-md" : "shadow-none"}
  `}
    >
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <div className="shrink-0">
            <Link href="/">
              {/* <SpeedcubingFullSVG size={170} textColor="black" /> */}
              <SpeedcubingFullAnimatedSVG
                animated={false}
                textColor="black"
                maxWidth={210}
              />
            </Link>
          </div>

          {/* Desktop Navigation - Hidden on mobile */}
          <div className="hidden md:flex md:space-x-8 items-center">
            {children}
          </div>

          {/* Mobile menu button */}
          <div className="md:hidden">
            <button
              onClick={toggleMenu}
              className="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
              aria-expanded="false"
            >
              <span className="sr-only">Open main menu</span>
              {/* Hamburger icon */}
              <svg
                className={`${isMenuOpen ? "hidden" : "block"} h-6 w-6`}
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
              {/* Close icon */}
              <svg
                className={`${isMenuOpen ? "block" : "hidden"} h-6 w-6`}
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Navigation Menu */}
      <div
        className={`
    md:hidden
    overflow-hidden
    
    ${isMenuOpen ? "max-h-96 opacity-100 transition-all duration-400 ease-in-out" : "max-h-0 opacity-0"}
  `}
      >
        <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 bg-white dark:bg-gray-800 shadow-lg">
          {React.Children.map(children, (child) => {
            if (!React.isValidElement(child)) return child;

            // Csak mobilon adjuk az onClick-et
            const props: any = {
              className:
                "block px-3 py-2 text-base font-medium text-gray-700 dark:text-gray-200 hover:text-gray-900 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-md",
            };
            if (isMenuOpen) {
              props.onClick = () => setIsMenuOpen(false);
            }

            return React.cloneElement(child, props);
          })}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
