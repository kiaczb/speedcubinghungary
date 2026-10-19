import type { IconType } from "react-icons";
import {
  FaBullseye,
  FaCalendarAlt,
  FaEnvelope,
  FaFacebook,
  FaGlobe,
  FaNewspaper,
  FaStopwatch,
  FaTrophy,
  FaUserFriends,
} from "react-icons/fa";

export type NavLink = {
  href: string;
  label: string;
  icon: IconType;
};

export type NavGroup = {
  label: string;
  links: NavLink[];
};

export type NavItem = NavLink | NavGroup;

export const isNavGroup = (item: NavItem): item is NavGroup => "links" in item;

export const NAV_ITEMS: NavItem[] = [
  {
    label: "Versenyek",
    links: [
      { href: "/tools/allrounder", label: "Hungarian All-Rounder", icon: FaTrophy },
      { href: "/tools/special", label: "Budapest Special", icon: FaStopwatch },
    ],
  },
  {
    label: "Közösség",
    links: [{ href: "/tools/prediction", label: "Tippverseny", icon: FaBullseye }],
  },
  { href: "/posts", label: "Hírek", icon: FaNewspaper },
  {
    label: "Egyesület",
    links: [
      { href: "/about", label: "Rólunk", icon: FaUserFriends },
      { href: "/contact", label: "Kapcsolat", icon: FaEnvelope },
    ],
  },
];

export const NAV_LINKS: NavLink[] = [
  { href: "/calendar", label: "Versenynaptár", icon: FaCalendarAlt },
  ...NAV_ITEMS.flatMap((item) => (isNavGroup(item) ? item.links : [item])),
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
