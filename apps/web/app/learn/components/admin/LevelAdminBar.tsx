"use client";
import { useState } from "react";
import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import { FaPen, FaPlus, FaRegTrashAlt } from "react-icons/fa";
import { useUser } from "@/app/components/UserProvider";
import Button from "@/app/components/ui/Button";
import { clientFetch } from "@/app/lib/api";
import { deleteNodeDescription, levelHref } from "../../lib/nodes";
import type { LearnLevel } from "../../types";

const NodeFormModal = dynamic(() => import("./NodeFormModal"), { ssr: false });
const AlgorithmFormModal = dynamic(() => import("./AlgorithmFormModal"), { ssr: false });
const ConfirmModal = dynamic(() => import("@/app/components/ui/ConfirmModal"), {
  ssr: false,
});

type Dialog = "create-node" | "create-algorithm" | "edit" | "delete";

type Props = {
  level: LearnLevel;
};

export default function LevelAdminBar({ level }: Props) {
  const { isAdmin } = useUser();
  const router = useRouter();
  const [dialog, setDialog] = useState<Dialog | null>(null);

  if (!isAdmin) return null;

  const { node, ancestors, children, algorithms } = level;
  const canAddNode = algorithms.length === 0;
  const canAddAlgorithm = node !== null && children.length === 0;
  const location = node
    ? [...ancestors, node].map((n) => n.title).join(" › ")
    : "Legfelső szint";
  const close = () => setDialog(null);

  const deleteCurrent = async () => {
    if (!node) return;
    await clientFetch(`/learn/nodes/${node.id}`, { method: "DELETE" });
    router.replace(levelHref(node.path.slice(0, -1)));
  };

  return (
    <>
      <div className="mt-8 flex flex-wrap items-center gap-2 rounded-card border border-line bg-surface p-3">
        {canAddNode && (
          <Button
            size="sm"
            onClick={() => setDialog("create-node")}
            icon={<FaPlus className="h-3 w-3" />}
          >
            Új elem
          </Button>
        )}
        {canAddAlgorithm && (
          <Button
            size="sm"
            variant={canAddNode ? "secondary" : "primary"}
            onClick={() => setDialog("create-algorithm")}
            icon={<FaPlus className="h-3 w-3" />}
          >
            Új algoritmus
          </Button>
        )}

        {node && (
          <div className="ml-auto flex flex-wrap gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setDialog("edit")}
              icon={<FaPen className="h-3 w-3" />}
            >
              Szerkesztés
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setDialog("delete")}
              icon={<FaRegTrashAlt className="h-3 w-3" />}
              className="hover:bg-brand-red/10 hover:text-brand-red"
            >
              Törlés
            </Button>
          </div>
        )}

        {canAddNode && canAddAlgorithm && (
          <p className="basis-full px-1 text-xs leading-relaxed text-subtle-foreground">
            Ez az elem még üres: kerülhetnek alá további elemek vagy algoritmusok —
            de a kettő együtt nem.
          </p>
        )}
      </div>

      {dialog === "create-node" && (
        <NodeFormModal
          mode="create"
          parentId={node?.id ?? null}
          locationLabel={location}
          onClose={close}
        />
      )}
      {dialog === "create-algorithm" && node && (
        <AlgorithmFormModal
          mode="create"
          nodeId={node.id}
          nodeTitle={node.title}
          onClose={close}
        />
      )}
      {dialog === "edit" && node && (
        <NodeFormModal mode="edit" node={node} isCurrentPage onClose={close} />
      )}
      {dialog === "delete" && node && (
        <ConfirmModal
          title="Elem törlése"
          description={deleteNodeDescription(node)}
          onConfirm={deleteCurrent}
          onClose={close}
        />
      )}
    </>
  );
}
