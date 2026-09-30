"use client";
import { FaLock, FaSpinner, FaUserShield } from "react-icons/fa";
import { useUser } from "@/app/components/UserProvider";
import Button from "@/app/components/ui/Button";
import StateMessage from "@/app/components/ui/StateMessage";
import AdminToolGrid from "./AdminToolGrid";

export default function AdminGate() {
  const { user, isAdmin, isLoading } = useUser();

  if (isLoading) {
    return (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Betöltés…"
        description="Jogosultságok ellenőrzése."
      />
    );
  }

  if (!user) {
    return (
      <StateMessage
        icon={<FaLock className="h-5 w-5" />}
        title="Bejelentkezés szükséges"
        description="Ehhez az oldalhoz be kell jelentkezned WCA-fiókoddal."
      >
        <Button href="/" variant="secondary" size="sm">
          Vissza a főoldalra
        </Button>
      </StateMessage>
    );
  }

  if (!isAdmin) {
    return (
      <StateMessage
        icon={<FaUserShield className="h-5 w-5" />}
        title="Nincs jogosultságod"
        description="Ez az oldal csak adminisztrátorok számára érhető el."
      >
        <Button href="/" variant="secondary" size="sm">
          Vissza a főoldalra
        </Button>
      </StateMessage>
    );
  }

  return <AdminToolGrid />;
}
