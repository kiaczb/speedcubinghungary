"use client";
import React, { createContext, useContext, useEffect, useState } from "react";
import { clientFetch, clientJson } from "../lib/api";

export interface MeResponse {
  user_id: number;
  wca_id: string;
  name: string;
  avatarURL?: string;
  roles: string[];
}

interface UserContextType {
  user: MeResponse | null;
  isLoading: boolean; // ÚJ: Töltési állapot
  logout: () => Promise<void>;
}

const UserContext = createContext<UserContextType | null>(null);

export const useUser = () => {
  const context = useContext(UserContext);
  if (!context) throw new Error("useUser must be used within a UserProvider");
  return context;
};

export const UserProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true); // Kezdetben töltünk

  useEffect(() => {
    const fetchMe = async () => {
      try {
        // A 401 itt VÁRT állapot: nincs bejelentkezve. Ezért nem naplózunk.
        setUser(await clientJson<MeResponse>("/me"));
      } catch {
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchMe();
  }, []);

  const logout = async () => {
    try {
      await clientFetch("/logout", { method: "POST" });
    } catch (error) {
      console.error("Logout failed", error);
    }
    // A helyi állapotot akkor is töröljük, ha a kérés elbukott — a süti
    // egy órán belül amúgy is lejár, és a felhasználó kijelentkezettnek hiszi
    // magát.
    setUser(null);
  };

  return (
    <UserContext.Provider value={{ user, isLoading, logout }}>
      {children}
    </UserContext.Provider>
  );
};