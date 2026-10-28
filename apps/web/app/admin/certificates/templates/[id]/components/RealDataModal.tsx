"use client";

import { useState } from "react";
import Modal from "@/app/components/ui/Modal";
import { clientJson } from "@/app/lib/api";
import CompetitionPicker from "../../../../stats/components/CompetitionPicker";
import type { RecipientList } from "../../../types";
import { certificateContexts } from "../../../lib/pdf/content";
import type { TemplateKind } from "../../../lib/template/model";
import type { PreviewSet } from "../../../lib/template/samples";

export default function RealDataModal({
  kind,
  onLoaded,
  onClose,
}: {
  kind: TemplateKind;
  onLoaded: (set: PreviewSet) => void;
  onClose: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const mode = kind === "placement" ? "podium" : "competitors";

  const load = async (competitionId: string) => {
    setLoading(true);
    setMessage(null);
    try {
      const params = new URLSearchParams({ competition: competitionId, mode });
      let list = await clientJson<RecipientList>(`/certificates/recipients?${params}`);
      if (list.recipients.length === 0 && mode === "competitors") {
        params.set("mode", "registrants");
        list = await clientJson<RecipientList>(`/certificates/recipients?${params}`);
      }
      if (list.recipients.length === 0) {
        setMessage("Ennél a versenynél még nincs címzett ehhez a típushoz (pl. még nincs lezárt döntő). Válassz másikat.");
        return;
      }
      onLoaded({ label: list.competition.name, contexts: certificateContexts(list.competition, list.mode, list.recipients) });
      onClose();
    } catch (error) {
      console.error("A verseny adatai nem kérhetők le:", error);
      setMessage("A verseny adatai most nem kérhetők le. Próbáld újra kicsit később.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal title="Előnézet valódi versenyadattal" onClose={onClose} size="lg" closeDisabled={loading}>
      <div className="flex flex-col gap-4 px-6 py-5">
        <p className="text-sm text-muted-foreground">
          A sablon nem kötődik versenyhez — ez csak az előnézethez tölti be egy verseny{" "}
          {kind === "placement" ? "dobogósait" : "versenyzőit"}, hogy lásd, minden név és adat elfér-e.
        </p>
        {message && <p className="text-sm text-brand-red">{message}</p>}
        {loading ? <p className="text-sm text-muted-foreground">Betöltés…</p> : <CompetitionPicker onSelect={load} />}
      </div>
    </Modal>
  );
}
