import { FaRegFolderOpen } from "react-icons/fa";
import PageHeader from "@/app/components/ui/PageHeader";
import StateMessage from "@/app/components/ui/StateMessage";
import { ALGORITHMS_DESCRIPTION } from "../lib/nodes";
import type { LearnLevel } from "../types";
import AlgorithmPanel from "./AlgorithmPanel";
import LevelAdminBar from "./admin/LevelAdminBar";
import LearnBreadcrumbs from "./LearnBreadcrumbs";
import NodeGrid from "./NodeGrid";

type Props = {
  level: LearnLevel;
};

export default function LevelView({ level }: Props) {
  const { node, ancestors, children, algorithms } = level;
  const isEmpty = children.length === 0 && algorithms.length === 0;

  return (
    <>
      <LearnBreadcrumbs ancestors={ancestors} node={node} />

      <PageHeader
        title={node?.title ?? "Algoritmusok"}
        description={node ? node.description : ALGORITHMS_DESCRIPTION}
        className="mt-6"
      />

      <LevelAdminBar level={level} />

      {children.length > 0 && <NodeGrid parentId={node?.id ?? null} nodes={children} />}

      {node && algorithms.length > 0 && (
        <AlgorithmPanel node={node} algorithms={algorithms} />
      )}

      {isEmpty && (
        <div className="mt-10">
          <StateMessage
            icon={<FaRegFolderOpen className="h-5 w-5" />}
            title={node ? "Ez a rész még üres" : "Még nincs tananyag"}
            description={
              node
                ? "Ide még nem került tartalom — nézz vissza később."
                : "Hamarosan itt találod a kirakókat, a módszereket és az algoritmusokat."
            }
          />
        </div>
      )}
    </>
  );
}
