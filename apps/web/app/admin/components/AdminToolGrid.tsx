import { ADMIN_TOOLS } from "../lib/tools";
import AdminToolTile from "./AdminToolTile";

export default function AdminToolGrid() {
  return (
    <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
      {ADMIN_TOOLS.map((tool) => (
        <AdminToolTile key={tool.id} tool={tool} />
      ))}
    </div>
  );
}
