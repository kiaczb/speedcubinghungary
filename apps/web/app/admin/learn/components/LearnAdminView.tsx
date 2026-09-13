"use client";
import { useCallback, useEffect, useState } from "react";
import {
  FaCheckCircle,
  FaExclamationTriangle,
  FaExternalLinkAlt,
  FaSpinner,
} from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import StateMessage from "@/app/components/ui/StateMessage";
import type { LearnEventNode } from "@/app/learn/types";
import {
  deleteAlgorithm,
  deleteCase,
  deleteEvent,
  fetchLearnTree,
} from "../lib/api";
import type { LearnFeedback, LearnModalState } from "../types";
import LearnColumn from "./LearnColumn";
import LearnEntityModal from "./LearnEntityModal";

// Hierarchikus admin felület: versenyszám → eset → algoritmus. A kiválasztás
// tölti fel a jobbra lévő hasábot, így a szülő mindig látszik, amikor gyereket
// veszünk fel. A kiválasztott elemeket a fából SZÁRMAZTATJUK (nem külön
// másolatból), így egy törlés vagy átnevezés után sem marad félreálló állapot.
export default function LearnAdminView() {
  const [tree, setTree] = useState<LearnEventNode[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const [selectedCaseId, setSelectedCaseId] = useState<string | null>(null);
  const [modal, setModal] = useState<LearnModalState | null>(null);
  const [feedback, setFeedback] = useState<LearnFeedback | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  // Kézi újratöltés mentés/törlés után — eseménykezelőből hívva.
  const refresh = useCallback(async () => {
    try {
      setTree(await fetchLearnTree());
      setLoadError(null);
    } catch (err) {
      setLoadError(
        err instanceof Error ? err.message : "A tartalom betöltése sikertelen.",
      );
    }
  }, []);

  // Az első betöltés a hatásban ígéret-visszahívással állít állapotot (nem az
  // effect törzsében) — ezt kéri a React Compiler lintje, és ez a `cancelled`
  // őr védi a közben lecsatolt komponenstől is.
  useEffect(() => {
    let cancelled = false;

    fetchLearnTree()
      .then((data) => {
        if (cancelled) return;
        setTree(data);
        setLoadError(null);
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setLoadError(
          error instanceof Error
            ? error.message
            : "A tartalom betöltése sikertelen.",
        );
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const selectedEvent =
    tree?.find((event) => event.id === selectedEventId) ?? null;
  const selectedCase =
    selectedEvent?.cases.find((item) => item.id === selectedCaseId) ?? null;

  const handleDelete = async (
    id: string,
    confirmation: string,
    remove: (id: string) => Promise<void>,
  ) => {
    if (!window.confirm(confirmation)) return;

    setBusyId(id);
    try {
      await remove(id);
      setFeedback({ type: "success", message: "A törlés megtörtént." });
      await refresh();
    } catch (err) {
      setFeedback({
        type: "error",
        message: err instanceof Error ? err.message : "A törlés sikertelen.",
      });
    } finally {
      setBusyId(null);
    }
  };

  if (loadError) {
    return (
      <StateMessage
        icon={<FaExclamationTriangle className="h-5 w-5" />}
        title="Nem sikerült betölteni"
        description={loadError}
      >
        <Button variant="secondary" size="sm" onClick={() => void refresh()}>
          Újrapróbálom
        </Button>
      </StateMessage>
    );
  }

  if (!tree) {
    return (
      <StateMessage
        icon={<FaSpinner className="h-5 w-5 animate-spin" />}
        title="Betöltés…"
        description="A tanulóanyag lekérése folyamatban."
      />
    );
  }

  return (
    <>
      {feedback && (
        <div
          role="status"
          className={`mb-6 flex items-start gap-2.5 rounded-control border px-3.5 py-2.5 text-sm ${
            feedback.type === "success"
              ? "border-brand-green/20 bg-brand-green/5 text-brand-green-strong"
              : "border-brand-red/20 bg-brand-red/5 text-brand-red-strong"
          }`}
        >
          {feedback.type === "success" ? (
            <FaCheckCircle aria-hidden className="mt-0.5 h-4 w-4 shrink-0" />
          ) : (
            <FaExclamationTriangle
              aria-hidden
              className="mt-0.5 h-4 w-4 shrink-0"
            />
          )}
          <span className="flex-1">{feedback.message}</span>
          <button
            type="button"
            onClick={() => setFeedback(null)}
            aria-label="Értesítés bezárása"
            className="-mr-1 -mt-0.5 text-lg leading-none opacity-70 transition-opacity hover:opacity-100"
          >
            ×
          </button>
        </div>
      )}

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
        <LearnColumn
          title="Versenyszámok"
          addLabel="Új"
          emptyLabel="Még nincs versenyszám. Kezdd egy létrehozásával."
          items={tree.map((event) => ({
            id: event.id,
            primary: event.name,
            meta: `${event.cases.length} eset`,
          }))}
          selectedId={selectedEventId}
          onSelect={(id) => {
            setSelectedEventId(id);
            setSelectedCaseId(null);
          }}
          onAdd={() => setModal({ kind: "event", entity: null })}
          onEdit={(id) => {
            const event = tree.find((item) => item.id === id);
            if (event) setModal({ kind: "event", entity: event });
          }}
          onDelete={(id) => {
            const event = tree.find((item) => item.id === id);
            if (!event) return;

            const algorithmCount = event.cases.reduce(
              (sum, item) => sum + item.algorithms.length,
              0,
            );
            void handleDelete(
              id,
              `Biztosan törlöd a(z) „${event.name}” versenyszámot? Ezzel törlődik alatta ${event.cases.length} eset és ${algorithmCount} algoritmus is. Ez nem vonható vissza.`,
              deleteEvent,
            );
          }}
          busyId={busyId}
        />

        <LearnColumn
          title="Esetek"
          addLabel="Új"
          parentHint={
            selectedEvent ? undefined : "Válassz versenyszámot a bal oldalon."
          }
          emptyLabel="Ehhez a versenyszámhoz még nincs eset."
          items={(selectedEvent?.cases ?? []).map((item) => ({
            id: item.id,
            primary: item.name,
            meta: `${item.algorithms.length} algoritmus`,
          }))}
          selectedId={selectedCaseId}
          onSelect={setSelectedCaseId}
          onAdd={() => {
            if (selectedEvent) {
              setModal({
                kind: "case",
                entity: null,
                eventId: selectedEvent.id,
              });
            }
          }}
          onEdit={(id) => {
            const item = selectedEvent?.cases.find((c) => c.id === id);
            if (item) {
              setModal({ kind: "case", entity: item, eventId: item.eventId });
            }
          }}
          onDelete={(id) => {
            const item = selectedEvent?.cases.find((c) => c.id === id);
            if (!item) return;

            void handleDelete(
              id,
              `Biztosan törlöd a(z) „${item.name}” esetet? Ezzel törlődik alatta ${item.algorithms.length} algoritmus is. Ez nem vonható vissza.`,
              deleteCase,
            );
          }}
          busyId={busyId}
        />

        <LearnColumn
          title="Algoritmusok"
          addLabel="Új"
          parentHint={selectedCase ? undefined : "Válassz esetet középen."}
          emptyLabel="Ehhez az esethez még nincs algoritmus."
          items={(selectedCase?.algorithms ?? []).map((algorithm) => ({
            id: algorithm.id,
            primary: algorithm.name,
            secondary: algorithm.notation,
            meta: algorithm.videoUrl ? "Videóval" : undefined,
          }))}
          onAdd={() => {
            if (selectedCase) {
              setModal({
                kind: "algorithm",
                entity: null,
                caseId: selectedCase.id,
              });
            }
          }}
          onEdit={(id) => {
            const algorithm = selectedCase?.algorithms.find((a) => a.id === id);
            if (algorithm) {
              setModal({
                kind: "algorithm",
                entity: algorithm,
                caseId: algorithm.caseId,
              });
            }
          }}
          onDelete={(id) => {
            const algorithm = selectedCase?.algorithms.find((a) => a.id === id);
            if (!algorithm) return;

            void handleDelete(
              id,
              `Biztosan törlöd a(z) „${algorithm.name}” algoritmust? Ez nem vonható vissza.`,
              deleteAlgorithm,
            );
          }}
          busyId={busyId}
        />
      </div>

      <div className="mt-6 flex flex-wrap items-center justify-between gap-3">
        <p className="text-xs leading-relaxed text-muted-foreground">
          A versenyszám törlése az alatta lévő eseteket és algoritmusokat is
          törli. A sorrend növekvő, azonos érték esetén név szerint rendezünk.
        </p>

        <Button
          href="/learn"
          variant="secondary"
          size="sm"
          icon={<FaExternalLinkAlt className="h-3 w-3" />}
        >
          Tanulófelület megnyitása
        </Button>
      </div>

      {modal && (
        <LearnEntityModal
          state={modal}
          tree={tree}
          onClose={() => setModal(null)}
          onSaved={(message) => {
            setModal(null);
            setFeedback({ type: "success", message });
            void refresh();
          }}
        />
      )}
    </>
  );
}
