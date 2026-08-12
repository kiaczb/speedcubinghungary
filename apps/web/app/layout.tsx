import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import LoginButton from "./components/LoginButton";
import { UserProvider } from "./components/UserProvider";
import { SITE_DESCRIPTION } from "./lib/site";

// A magyar ő/ű ékezetekhez a latin-ext subset is kell, nem csak latin.
const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin", "latin-ext"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin", "latin-ext"],
});

export const metadata: Metadata = {
  title: {
    default: "Speedcubing Hungary — a magyar kockázó közösség",
    template: "%s | Speedcubing Hungary",
  },
  description: SITE_DESCRIPTION,
  openGraph: {
    title: "Speedcubing Hungary",
    description: SITE_DESCRIPTION,
    locale: "hu_HU",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="hu">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased flex min-h-screen flex-col bg-canvas`}
      >
        <a
          href="#main"
          className="sr-only rounded-control bg-foreground px-4 py-2 text-sm font-semibold text-background focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-60"
        >
          Ugrás a tartalomra
        </a>

        <UserProvider>
          <Navbar action={<LoginButton />} />
          <main id="main" className="flex-1">
            {children}
          </main>
        </UserProvider>
        <Footer />
      </body>
    </html>
  );
}
