"use client";
import React from "react";
import { usePathname } from "next/navigation";
import { clientApiUrl } from "../lib/api";
import { useUser } from "./UserProvider";
import Button from "./ui/Button";

const LoginButton = () => {
  const { user, isAdmin, isLoading, logout } = useUser();
  const pathname = usePathname();

  const handleLogin = () => {
    window.location.href = `${clientApiUrl()}/login?redirect=${encodeURIComponent(pathname || "/")}`;
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
