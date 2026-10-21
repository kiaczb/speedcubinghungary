"use client";

import { useState } from "react";
import { FaKey } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import { clientApiUrl, clientFetch } from "@/app/lib/api";
import type { AgedPerson, AgeStats } from "../types";
import {
  KeyFigures,
  PersonName,
  Rows,
  StatBlock,
  StatSection,
} from "./Blocks";

const DESCRIPTION =
  "A WCA születési dátumai alapján, a verseny első napjára számolva. A dátumokat nem tároljuk és nem mutatjuk, csak az életkort.";

function formatAge(person: AgedPerson): string {
  return person.days > 0
    ? `${person.years} év ${person.days} nap`
    : `${person.years} év`;
}

function ageRows(people: AgedPerson[]) {
  return people.map((aged) => ({
    key: aged.person.wcaId ?? aged.person.name,
    label: <PersonName person={aged.person} />,
    value: formatAge(aged),
  }));
}

export default function AgesSection({
  competitionId,
  ages,
  onChanged,
}: {
  competitionId: string;
  ages: AgeStats;
  onChanged: () => void;
}) {
  const [revoking, setRevoking] = useState(false);

  const authorizeHref = `${clientApiUrl()}/wca/authorize?redirect=${encodeURIComponent(
    `/admin/stats?competition=${competitionId}`,
  )}`;

  const revoke = async () => {
    setRevoking(true);
    try {
      await clientFetch("/wca/grant", { method: "DELETE" });
      onChanged();
    } catch (error) {
      console.error("A WCA-hozzáférés visszavonása sikertelen:", error);
    } finally {
      setRevoking(false);
    }
  };

  if (ages.status === "no-grant") {
    return (
      <StatSection title="Életkorok" description={DESCRIPTION}>
        <div className="flex max-w-2xl flex-col items-start gap-4">
          <p className="text-sm leading-relaxed text-muted-foreground">
            A legfiatalabb és a legidősebb versenyzőhöz a születési dátumok
            kellenek. Ezeket a WCA csak a verseny szervezőinek és
            delegáltjainak adja ki, és ehhez egyszer engedélyezned kell a
            versenykezelési hozzáférést a WCA-fiókodhoz.
          </p>
          <Button
            href={authorizeHref}
            variant="secondary"
            size="sm"
            icon={<FaKey className="h-3 w-3" />}
          >
            WCA-hozzáférés engedélyezése
          </Button>
        </div>
      </StatSection>
    );
  }

  if (ages.status !== "ready") {
    return (
      <StatSection title="Életkorok" description={DESCRIPTION}>
        <p className="max-w-2xl text-sm leading-relaxed text-muted-foreground">
          {ages.status === "forbidden"
            ? "Ennek a versenynek nem vagy szervezője vagy delegáltja a WCA-n, ezért a születési dátumokat nem kapjuk meg."
            : "A születési dátumok most nem kérhetők le a WCA-tól. Próbáld újra kicsit később."}
        </p>
      </StatSection>
    );
  }

  return (
    <StatSection title="Életkorok" description={DESCRIPTION}>
      <KeyFigures
        figures={[
          { label: "Átlagéletkor", value: `${ages.average.toLocaleString("hu-HU")} év` },
          { label: "Medián", value: `${ages.median.toLocaleString("hu-HU")} év` },
          {
            label: "Legfiatalabb",
            value: ages.youngest[0] ? `${ages.youngest[0].years} év` : "–",
          },
          {
            label: "Legidősebb",
            value: ages.oldest[0] ? `${ages.oldest[0].years} év` : "–",
          },
          { label: "Ismert életkor", value: ages.known, sub: "versenyző" },
        ]}
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <StatBlock title="Legfiatalabbak">
          <Rows empty="Nincs adat." rows={ageRows(ages.youngest)} />
        </StatBlock>
        <StatBlock title="Legidősebbek">
          <Rows empty="Nincs adat." rows={ageRows(ages.oldest)} />
        </StatBlock>
      </div>

      <div>
        <Button variant="ghost" size="sm" onClick={revoke} disabled={revoking}>
          WCA-hozzáférés visszavonása
        </Button>
      </div>
    </StatSection>
  );
}
