"use client";
import { useId, useState } from "react";
import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import {
  FaArrowDown,
  FaArrowUp,
  FaChevronDown,
  FaListOl,
  FaPen,
  FaRegTrashAlt,
} from "react-icons/fa";
import Card from "@/app/components/Cards/Card";
import { useUser } from "@/app/components/UserProvider";
import { clientFetch } from "@/app/lib/api";
import { moveItem } from "../lib/nodes";
import type { LearnAlgorithm, LearnNode } from "../types";
import AdminIconButton from "./admin/AdminIconButton";
import AlgorithmItem from "./AlgorithmItem";

const AlgorithmFormModal = dynamic(() => import("./admin/AlgorithmFormModal"), {
  ssr: false,
});
const ConfirmModal = dynamic(() => import("@/app/components/ui/ConfirmModal"), {
  ssr: false,
});

type Props = {
  node: LearnNode;
  algorithms: LearnAlgorithm[];
};

export default function AlgorithmPanel({ node, algorithms }: Props) {
  const { isAdmin } = useUser();
  const router = useRouter();
  const bodyId = useId();
  const [open, setOpen] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<LearnAlgorithm | null>(null);
  const [deleting, setDeleting] = useState<LearnAlgorithm | null>(null);

  const move = async (index: number, delta: -1 | 1) => {
    setError(null);
    setBusy(true);
    try {
      await clientFetch(`/learn/nodes/${node.id}/algorithms/order`, {
        method: "PUT",
        json: { ids: moveItem(algorithms.map((a) => a.id), index, delta) },
      });
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
    } finally {
      setBusy(false);
    }
  };

  const remove = async () => {
    if (!deleting) return;
    await clientFetch(`/learn/algorithms/${deleting.id}`, { method: "DELETE" });
    router.refresh();
  };

  return (
    <section className="mt-10">
      <Card className="overflow-hidden">
        <h2>
          <button
            type="button"
            onClick={() => setOpen((value) => !value)}
            aria-expanded={open}
            aria-controls={bodyId}
            className="flex w-full items-center gap-4 px-5 py-4 text-left transition-colors duration-200 ease-smooth hover:bg-surface-muted/60 sm:px-6"
          >
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-control bg-surface-muted text-foreground">
              <FaListOl className="h-4 w-4" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="block text-lg font-semibold text-foreground">
                Algoritmusok
              </span>
              <span className="tabular block font-mono text-xs text-subtle-foreground">
                {node.title} · {algorithms.length} algoritmus
              </span>
            </span>
            <FaChevronDown
              className={`h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform duration-200 ease-smooth ${
                open ? "rotate-180" : ""
              }`}
              aria-hidden
            />
          </button>
        </h2>

        <div id={bodyId} hidden={!open} className="border-t border-line">
          {error && (
            <p
              role="alert"
              className="mx-5 mt-4 rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong sm:mx-6"
            >
              Nem sikerült az átrendezés: {error}
            </p>
          )}

          <ol className="divide-y divide-line">
            {algorithms.map((algorithm, index) => (
              <AlgorithmItem
                key={algorithm.id}
                algorithm={algorithm}
                index={index}
                caseTitle={node.title}
                actions={
                  isAdmin && (
                    <div className="-ml-2 mt-3 flex items-center">
                      <AdminIconButton
                        label={`${index + 1}. algoritmus előrébb`}
                        onClick={() => move(index, -1)}
                        disabled={busy || index === 0}
                      >
                        <FaArrowUp className="h-3 w-3" />
                      </AdminIconButton>
                      <AdminIconButton
                        label={`${index + 1}. algoritmus hátrébb`}
                        onClick={() => move(index, 1)}
                        disabled={busy || index === algorithms.length - 1}
                      >
                        <FaArrowDown className="h-3 w-3" />
                      </AdminIconButton>
                      <AdminIconButton
                        label={`${index + 1}. algoritmus szerkesztése`}
                        onClick={() => setEditing(algorithm)}
                      >
                        <FaPen className="h-3 w-3" />
                      </AdminIconButton>
                      <AdminIconButton
                        label={`${index + 1}. algoritmus törlése`}
                        onClick={() => setDeleting(algorithm)}
                        danger
                      >
                        <FaRegTrashAlt className="h-3.5 w-3.5" />
                      </AdminIconButton>
                    </div>
                  )
                }
              />
            ))}
          </ol>
        </div>
      </Card>

      {editing && (
        <AlgorithmFormModal
          mode="edit"
          algorithm={editing}
          nodeTitle={node.title}
          onClose={() => setEditing(null)}
        />
      )}
      {deleting && (
        <ConfirmModal
          title="Algoritmus törlése"
          description={
            <>
              Biztosan törlöd ezt az algoritmust:{" "}
              <span className="font-mono text-foreground">{deleting.notation}</span>?
              Ez nem vonható vissza.
            </>
          }
          onConfirm={remove}
          onClose={() => setDeleting(null)}
        />
      )}
    </section>
  );
}
