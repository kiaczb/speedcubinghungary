"use client";
import React from "react";
import { usePathname } from "next/navigation";
import { clientApiUrl } from "../lib/api";
import { useUser } from "./UserProvider";
import Button from "./ui/Button";

const LoginButton = () => {
  const { user, isLoading, logout } = useUser();
  const pathname = usePathname();

  // A bejelentkezést a BACKEND indítja (GET /login), nem itt építjük a WCA
  // authorize URL-t.
  //
  // Miért: az OAuth `state` a folyamat CSRF-tokenje, ezért kitalálhatatlan,
  // szerveroldalon generált nonce-nak kell lennie, amit HttpOnly sütiben tárolunk
  // — böngészőből ezt nem lehet biztonságosan megtenni. Ráadásul így a
  // client_id / redirect_uri / scope hármas egyetlen helyen (a backend
  // configjában) él, nem csúszhat el a build időben beégetett NEXT_PUBLIC_*
  // értékek és a backend WCA_* értékei között.
  const handleLogin = () => {
    const loginUrl =
      process.env.NEXT_PUBLIC_BACKEND_LOGIN || `${clientApiUrl()}/login`;

    window.location.href = `${loginUrl}?redirect=${encodeURIComponent(pathname || "/")}`;
  };

  if (isLoading) {
    return (
      <div
        aria-hidden
        className="h-9 w-32 animate-pulse rounded-control bg-surface-muted"
      />
    );
  }

  if (user) {
    const isAdmin = user.roles.includes("admin");

    return (
      <div className="flex items-center gap-2">
        {isAdmin && (
          <Button href="/admin" variant="ghost" size="sm">
            Admin
          </Button>
        )}
        <Button
          variant="secondary"
          size="sm"
          onClick={logout}
          title={`Kijelentkezés — ${user.name}`}
          icon={
            user.avatarURL ? (
              /* next/image helyett sima img: a kép a WCA-ról jön (külső domain),
                 remotePatterns config nélkül így szolgálható ki. */
              /* eslint-disable-next-line @next/next/no-img-element */
              <img
                src={user.avatarURL}
                alt=""
                className="-ml-1 h-6 w-6 rounded-full object-cover ring-1 ring-line"
              />
            ) : undefined
          }
        >
          <span className="max-w-28 truncate">{user.name}</span>
          <span className="text-muted-foreground">· Kilépés</span>
        </Button>
      </div>
    );
  }

  return (
    <Button size="sm" onClick={handleLogin}>
      Bejelentkezés WCA-fiókkal
    </Button>
  );
};

export default LoginButton;
