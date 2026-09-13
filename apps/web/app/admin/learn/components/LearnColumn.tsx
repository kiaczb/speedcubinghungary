"use client";
import { FaPen, FaPlus, FaRegTrashAlt } from "react-icons/fa";
import Button from "@/app/components/ui/Button";

export type ColumnItem = {
  id: string;
  primary: string;
  secondary?: string;
  /** Rövid számláló a sor jobb szélén (pl. "3 eset"). */
  meta?: string;
};

type Props = {
  title: string;
  /** Ha nincs kiválasztott szülő, a hasáb ezt írja ki és nem lehet felvenni. */
  parentHint?: string;
  emptyLabel: string;
  addLabel: string;
  items: ColumnItem[];
  selectedId?: string | null;
  onSelect?: (id: string) => void;
  onAdd: () => void;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
  /** Az éppen mentés/törlés alatt álló elem — a sor gombjai letiltva. */
  busyId?: string | null;
};

// A három szint UI-ja azonos, csak a tartalom más — egy hasáb, nem három
// külön komponens. A hierarchia balról jobbra olvasható: versenyszám → eset →
// algoritmus, és a kiválasztás tölti fel a jobbra lévő hasábot.
export default function LearnColumn({
  title,
  parentHint,
  emptyLabel,
  addLabel,
  items,
  selectedId,
  onSelect,
  onAdd,
  onEdit,
  onDelete,
  busyId,
}: Props) {
  const disabled = Boolean(parentHint);

  return (
    <section className="flex min-h-[18rem] flex-col rounded-card border border-line bg-surface shadow-card">
      <div className="flex items-center justify-between gap-3 border-b border-line px-4 py-3">
        <h2 className="text-sm font-semibold text-foreground">
          {title}
          {!disabled && items.length > 0 && (
            <span className="tabular ml-2 font-mono text-xs font-normal text-subtle-foreground">
              {items.length}
            </span>
          )}
        </h2>

        <Button
          size="sm"
          variant="secondary"
          onClick={onAdd}
          disabled={disabled}
          icon={<FaPlus className="h-3 w-3" />}
        >
          {addLabel}
        </Button>
      </div>

      <div className="flex-1 p-3">
        {disabled ? (
          <p className="px-2 py-8 text-center text-sm text-subtle-foreground">
            {parentHint}
          </p>
        ) : items.length === 0 ? (
          <p className="px-2 py-8 text-center text-sm text-subtle-foreground">
            {emptyLabel}
          </p>
        ) : (
          <ul className="flex flex-col gap-1.5">
            {items.map((item) => {
              const isSelected = item.id === selectedId;
              const isBusy = item.id === busyId;

              return (
                <li key={item.id}>
                  <div
                    className={`flex items-center gap-2 rounded-control border px-3 py-2 transition-colors duration-200 ease-smooth ${
                      isSelected
                        ? "border-foreground bg-surface-muted"
                        : "border-transparent hover:border-line hover:bg-surface-muted"
                    }`}
                  >
                    {/* Kiválasztható szinten a sor törzse gomb (a jobbra lévő
                        hasábot tölti fel), levélszinten sima szöveg. */}
                    {onSelect ? (
                      <button
                        type="button"
                        onClick={() => onSelect(item.id)}
                        aria-pressed={isSelected}
                        className="min-w-0 flex-1 text-left"
                      >
                        <ItemBody item={item} />
                      </button>
                    ) : (
                      <div className="min-w-0 flex-1">
                        <ItemBody item={item} />
                      </div>
                    )}

                    <div className="flex shrink-0 items-center gap-1">
                      <IconButton
                        label={`${item.primary} szerkesztése`}
                        onClick={() => onEdit(item.id)}
                        disabled={isBusy}
                      >
                        <FaPen className="h-3 w-3" />
                      </IconButton>
                      <IconButton
                        label={`${item.primary} törlése`}
                        onClick={() => onDelete(item.id)}
                        disabled={isBusy}
                        danger
                      >
                        <FaRegTrashAlt className="h-3 w-3" />
                      </IconButton>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </section>
  );
}

function ItemBody({ item }: { item: ColumnItem }) {
  return (
    <>
      <p className="truncate text-sm font-medium text-foreground">
        {item.primary}
      </p>
      {item.secondary && (
        <p className="tabular truncate font-mono text-xs text-muted-foreground">
          {item.secondary}
        </p>
      )}
      {item.meta && (
        <p className="tabular font-mono text-xs text-subtle-foreground">
          {item.meta}
        </p>
      )}
    </>
  );
}

type IconButtonProps = {
  label: string;
  onClick: () => void;
  disabled?: boolean;
  danger?: boolean;
  children: React.ReactNode;
};

function IconButton({
  label,
  onClick,
  disabled,
  danger,
  children,
}: IconButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
      title={label}
      className={`inline-flex h-7 w-7 items-center justify-center rounded-control transition-colors duration-200 ease-smooth disabled:pointer-events-none disabled:opacity-40 ${
        danger
          ? "text-muted-foreground hover:bg-brand-red/10 hover:text-brand-red"
          : "text-muted-foreground hover:bg-surface hover:text-foreground"
      }`}
    >
      {children}
    </button>
  );
}
