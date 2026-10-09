"use client";
import { useState } from "react";
import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import { FaArrowLeft, FaArrowRight, FaPen, FaRegTrashAlt } from "react-icons/fa";
import { useUser } from "@/app/components/UserProvider";
import { clientFetch } from "@/app/lib/api";
import { deleteNodeDescription, moveItem } from "../lib/nodes";
import type { LearnNode } from "../types";
import AdminIconButton from "./admin/AdminIconButton";
import NodeCard from "./NodeCard";

const NodeFormModal = dynamic(() => import("./admin/NodeFormModal"), { ssr: false });
const ConfirmModal = dynamic(() => import("@/app/components/ui/ConfirmModal"), {
  ssr: false,
});

type Props = {
  parentId: string | null;
  nodes: LearnNode[];
};

export default function NodeGrid({ parentId, nodes }: Props) {
  const { isAdmin } = useUser();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<LearnNode | null>(null);
  const [deleting, setDeleting] = useState<LearnNode | null>(null);
  const showMedia = nodes.some((n) => n.icon || n.imageUrl);

  const move = async (index: number, delta: -1 | 1) => {
    setError(null);
    setBusy(true);
    try {
      await clientFetch("/learn/nodes/order", {
        method: "PUT",
        json: { parentId, ids: moveItem(nodes.map((n) => n.id), index, delta) },
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
    await clientFetch(`/learn/nodes/${deleting.id}`, { method: "DELETE" });
    router.refresh();
  };

  return (
    <section className="mt-10">
      {error && (
        <p
          role="alert"
          className="mb-4 rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong"
        >
          Nem sikerült az átrendezés: {error}
        </p>
      )}

      <ul
        className={`grid grid-cols-2 gap-3 sm:gap-4 ${
          showMedia
            ? "sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
            : "md:grid-cols-3 lg:grid-cols-4"
        }`}
      >
        {nodes.map((node, index) => (
          <li key={node.id}>
            <NodeCard
              node={node}
              showMedia={showMedia}
              actions={
                isAdmin && (
                  <div className="-mx-1 mt-1 flex items-center border-t border-line pt-2">
                    <AdminIconButton
                      label={`${node.title} előrébb`}
                      onClick={() => move(index, -1)}
                      disabled={busy || index === 0}
                    >
                      <FaArrowLeft className="h-3 w-3" />
                    </AdminIconButton>
                    <AdminIconButton
                      label={`${node.title} hátrébb`}
                      onClick={() => move(index, 1)}
                      disabled={busy || index === nodes.length - 1}
                    >
                      <FaArrowRight className="h-3 w-3" />
                    </AdminIconButton>
                    <span className="flex-1" />
                    <AdminIconButton
                      label={`${node.title} szerkesztése`}
                      onClick={() => setEditing(node)}
                    >
                      <FaPen className="h-3 w-3" />
                    </AdminIconButton>
                    <AdminIconButton
                      label={`${node.title} törlése`}
                      onClick={() => setDeleting(node)}
                      danger
                    >
                      <FaRegTrashAlt className="h-3.5 w-3.5" />
                    </AdminIconButton>
                  </div>
                )
              }
            />
          </li>
        ))}
      </ul>

      {editing && (
        <NodeFormModal
          mode="edit"
          node={editing}
          isCurrentPage={false}
          onClose={() => setEditing(null)}
        />
      )}
      {deleting && (
        <ConfirmModal
          title="Elem törlése"
          description={deleteNodeDescription(deleting)}
          onConfirm={remove}
          onClose={() => setDeleting(null)}
        />
      )}
    </section>
  );
}
