"use client";
import React from "react";
import { usePathname } from "next/navigation";
import { useUser } from "./UserProvider";
import Button from "./ui/Button";

const LoginButton = () => {
  const { user, isLoading, logout } = useUser();
  const pathname = usePathname();

  const handleLogin = () => {
    const clientId = process.env.NEXT_PUBLIC_WCA_CLIENT_ID;
    const redirectUri = encodeURIComponent(
      process.env.NEXT_PUBLIC_WCA_REDIRECT_URI ||
        "http://localhost:8080/login/callback",
    );

    const state = encodeURIComponent(pathname || "/");

    const wcaAuthUrl =
      `https://www.worldcubeassociation.org/oauth/authorize` +
      `?client_id=${clientId}` +
      `&redirect_uri=${redirectUri}` +
      `&response_type=code` +
      `&scope=public+email+manage_competitions` +
      `&state=${state}`;

    window.location.href = wcaAuthUrl;
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
