import type { IconType } from "react-icons";
import { FaEnvelope, FaFacebook, FaGlobe } from "react-icons/fa";

type NavLink = {
  href: string;
  label: string;
};

export const NAV_LINKS: NavLink[] = [
  { href: "/calendar", label: "Versenynaptár" },
  { href: "/tools", label: "Eszközök" },
  { href: "/learn", label: "Tanulás" },
  { href: "/posts", label: "Hírek" },
  { href: "/about", label: "Rólunk" },
  { href: "/contact", label: "Kapcsolat" },
];

type ContactItem = {
  icon: IconType;
  label: string;
  href: string;
};

export const CONTACT_ITEMS: ContactItem[] = [
  {
    icon: FaEnvelope,
    label: "info@speedcubinghungary.hu",
    href: "mailto:info@speedcubinghungary.hu",
  },
  {
    icon: FaGlobe,
    label: "www.speedcubinghungary.hu",
    href: "https://www.speedcubinghungary.hu",
  },
  {
    icon: FaFacebook,
    label: "SpeedcubingHungary",
    href: "https://www.facebook.com/SpeedcubingHungary",
  },
];

export const SITE_DESCRIPTION =
  "A Speedcubing Hungary Egyesület célja a kockázás népszerűsítése és WCA regionális szervezetként hivatalos WCA versenyek szervezése.";
