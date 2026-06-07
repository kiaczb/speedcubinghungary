"use client";
import React from "react";
import { usePathname } from "next/navigation";
import { useUser } from "./UserProvider";

const LoginButton = () => {
  const { user, isLoading, logout } = useUser();
  const pathname = usePathname();

  const handleLogin = () => {
    const clientId = process.env.NEXT_PUBLIC_WCA_CLIENT_ID;
    // Ezt is az .env-ből olvassuk fel!
    const redirectUri = encodeURIComponent(
      process.env.NEXT_PUBLIC_WCA_REDIRECT_URI || "http://localhost:8080/login/callback"
    );

    const state = encodeURIComponent(pathname || "/");
    
    const wcaAuthUrl =
      `https://www.worldcubeassociation.org/oauth/authorize` +
      `?client_id=${clientId}` +
      `&redirect_uri=${redirectUri}` +
      `&response_type=code` +
      `&scope=public+email+manage_competitions`+
      `&state=${state}`;

    window.location.href = wcaAuthUrl;
  };

  if (isLoading) {
    return <div className="p-3 text-gray-500">Betöltés...</div>;
  }

  if (user) {
    return (
      <div>
        <button
          onClick={logout}
          className="p-3 bg-green-500 rounded-md text-white text-center flex items-center justify-between hover:bg-green-600 transition"
        >
          <span>Kijelentkezés ({user.name})</span>
        </button>
      </div>
    );
  }

  return (
    <div className="p-3 bg-blue-500 rounded-md text-white text-center hover:bg-blue-600 transition inline-block">
      <button onClick={handleLogin}>Login with WCA</button>
    </div>
  );
};

export default LoginButton;