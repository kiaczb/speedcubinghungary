// app/components/UserProvider.tsx
"use client";
import React, { createContext, useContext, useEffect, useState } from "react";

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

  // Környezeti változó használata a backend URL-hez (fallback a localhost-ra)
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  useEffect(() => {
    const fetchMe = async () => {
      try {
        const res = await fetch(`${API_URL}/me`, {
          credentials: "include", // EZ A KULCS! Ez küldi el a sütit.
        });
        if (!res.ok) throw new Error("Not logged in");
        
        const data: MeResponse = await res.json();
        setUser(data);
      } catch {
        setUser(null);
      } finally {
        setIsLoading(false); // Befejeztük a lekérést
      }
    };
    
    fetchMe();
  }, [API_URL]);

  const logout = async () => {
    try {
      await fetch(`${API_URL}/logout`, { 
        method: "POST", 
        credentials: "include" 
      });
      setUser(null);
    } catch (error) {
      console.error("Logout failed", error);
    }
  };

  return (
    <UserContext.Provider value={{ user, isLoading, logout }}>
      {children}
    </UserContext.Provider>
  );
};