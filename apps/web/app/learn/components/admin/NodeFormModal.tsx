"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import "@cubing/icons";
import Button from "@/app/components/ui/Button";
import Field, { controlClass } from "@/app/components/ui/Field";
import Modal from "@/app/components/ui/Modal";
import { clientFetch, clientJson } from "@/app/lib/api";
import { EVENT_ICONS, UNOFFICIAL_ICONS, iconLabel } from "../../lib/icons";
import { levelHref } from "../../lib/nodes";
import type { LearnNode } from "../../types";

type Props =
  | {
      mode: "create";
      parentId: string | null;
      locationLabel: string;
      onClose: () => void;
    }
  | {
      mode: "edit";
      node: LearnNode;
      isCurrentPage: boolean;
      onClose: () => void;
    };

type TreeState =
  | { status: "loading" }
  | { status: "error" }
  | { status: "ready"; nodes: LearnNode[] };

const ROOT_VALUE = "";

function isWithin(candidate: LearnNode, node: LearnNode): boolean {
  return node.path.every((slug, i) => candidate.path[i] === slug);
}

export default function NodeFormModal(props: Props) {
  const router = useRouter();
  const editing = props.mode === "edit" ? props.node : null;

  const [title, setTitle] = useState(editing?.title ?? "");
  const [slug, setSlug] = useState(editing?.slug ?? "");
  const [description, setDescription] = useState(editing?.description ?? "");
  const [icon, setIcon] = useState(editing?.icon ?? "");
  const [imageUrl, setImageUrl] = useState(editing?.imageUrl ?? "");
  const [parentId, setParentId] = useState<string | null>(
    editing ? editing.parentId : props.mode === "create" ? props.parentId : null,
  );
  const [tree, setTree] = useState<TreeState>({ status: "loading" });
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const titleRef = useRef<HTMLInputElement>(null);
  const editingId = editing?.id;

  useEffect(() => {
    if (!editingId) return;
    const controller = new AbortController();

    clientJson<LearnNode[]>("/learn/tree", { signal: controller.signal })
      .then((nodes) => setTree({ status: "ready", nodes }))
      .catch(() => {
        if (!controller.signal.aborted) setTree({ status: "error" });
      });

    return () => controller.abort();
  }, [editingId]);

  const parentOptions =
    editing && tree.status === "ready"
      ? tree.nodes.filter(
          (candidate) =>
            candidate.algorithmCount === 0 && !isWithin(candidate, editing),
        )
      : [];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!title.trim()) {
      setError("Adj címet az elemnek.");
      return;
    }

    const body = {
      parentId,
      title: title.trim(),
      slug: slug.trim(),
      description: description.trim(),
      icon,
      imageUrl: imageUrl.trim(),
    };

    setSubmitting(true);
    try {
      if (props.mode === "create") {
        await clientFetch("/learn/nodes", { method: "POST", json: body });
        router.refresh();
      } else {
        const saved = await clientJson<LearnNode>(`/learn/nodes/${props.node.id}`, {
          method: "PUT",
          json: body,
        });
        const href = levelHref(saved.path);
        if (props.isCurrentPage && href !== levelHref(props.node.path)) {
          router.replace(href);
        } else {
          router.refresh();
        }
      }
      props.onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ismeretlen hiba történt.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={editing ? `„${editing.title}” szerkesztése` : "Új elem"}
      onClose={props.onClose}
      size="lg"
      closeDisabled={submitting}
      initialFocusRef={titleRef}
    >
      <form onSubmit={handleSubmit} className="space-y-5 px-6 py-5">
        <Field
          label="Cím"
          htmlFor="learn-node-title"
          hint="Ez jelenik meg a kártyán, pl. „3x3”, „CFOP”, „OLL 24”."
        >
          <input
            id="learn-node-title"
            ref={titleRef}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={100}
            className={controlClass}
          />
        </Field>

        <Field
          label="Azonosító az URL-ben"
          htmlFor="learn-node-slug"
          optional
          hint={
            editing
              ? "Ha megváltoztatod, a korábbi linkek nem működnek tovább. Ékezet nélkül, kisbetűvel, kötőjellel tagolva mentjük."
              : "Üresen hagyva a címből készül (pl. „OLL 24” → oll-24). Ékezet nélkül, kisbetűvel, kötőjellel tagolva mentjük."
          }
        >
          <input
            id="learn-node-slug"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="automatikus"
            maxLength={64}
            spellCheck={false}
            autoCapitalize="off"
            className={`${controlClass} font-mono`}
          />
        </Field>

        <Field
          label="Leírás"
          htmlFor="learn-node-description"
          optional
          hint="Rövid magyarázat a kártyán és az oldal tetején."
        >
          <textarea
            id="learn-node-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            maxLength={1000}
            className={controlClass}
          />
        </Field>

        <Field
          label="Kép URL"
          htmlFor="learn-node-image"
          optional
          hint={
            <>
              Egy https:// kép címe; ha van, az ikon helyett ez jelenik meg. Esetábrát
              pl. a{" "}
              <a
                href="https://visualcube.api.cubing.net/"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-foreground underline underline-offset-2"
              >
                VisualCube
              </a>{" "}
              generál bármelyik algoritmusból. A képfeltöltés később jön.
            </>
          }
        >
          <div className="flex items-center gap-3">
            <span
              className="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-control bg-surface-muted"
              aria-hidden
            >
              {imageUrl.trim().startsWith("https://") ? (
                /* eslint-disable-next-line @next/next/no-img-element */
                <img src={imageUrl.trim()} alt="" className="h-full w-full object-contain" />
              ) : (
                <span className="text-subtle-foreground">—</span>
              )}
            </span>
            <input
              id="learn-node-image"
              type="url"
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
              placeholder="https://…"
              maxLength={2000}
              spellCheck={false}
              className={controlClass}
            />
          </div>
        </Field>

        <Field
          label="Ikon"
          htmlFor="learn-node-icon"
          optional
          hint="A kártya képe. Ha egy szinten bármelyik elemnek van ikonja, a többi kártya a cím kezdőbetűjét kapja helyette."
        >
          <div className="flex items-center gap-3">
            <span
              className="flex h-11 w-11 shrink-0 items-center justify-center rounded-control bg-surface-muted text-foreground"
              aria-hidden
            >
              {icon ? (
                <span className={`cubing-icon ${icon} text-2xl leading-none`} />
              ) : (
                <span className="text-subtle-foreground">—</span>
              )}
            </span>
            <select
              id="learn-node-icon"
              value={icon}
              onChange={(e) => setIcon(e.target.value)}
              className={controlClass}
            >
              <option value="">Nincs ikon</option>
              <optgroup label="Hivatalos versenyszámok">
                {EVENT_ICONS.map((id) => (
                  <option key={id} value={id}>
                    {iconLabel(id)}
                  </option>
                ))}
              </optgroup>
              <optgroup label="Nem hivatalos">
                {UNOFFICIAL_ICONS.map((id) => (
                  <option key={id} value={id}>
                    {iconLabel(id)}
                  </option>
                ))}
              </optgroup>
            </select>
          </div>
        </Field>

        {props.mode === "create" ? (
          <p className="text-sm text-muted-foreground">
            Hely:{" "}
            <span className="font-medium text-foreground">{props.locationLabel}</span>
          </p>
        ) : (
          <Field
            label="Hely a hierarchiában"
            htmlFor="learn-node-parent"
            hint={
              tree.status === "error"
                ? "Nem sikerült betölteni a hierarchiát — az elem a helyén marad."
                : "Áthelyezéskor a teljes tartalma vele költözik. Algoritmust tartalmazó elem alá nem kerülhet."
            }
          >
            <select
              id="learn-node-parent"
              value={parentId ?? ROOT_VALUE}
              onChange={(e) => setParentId(e.target.value === ROOT_VALUE ? null : e.target.value)}
              disabled={tree.status !== "ready"}
              className={controlClass}
            >
              <option value={ROOT_VALUE}>Legfelső szint</option>
              {tree.status === "ready"
                ? parentOptions.map((option) => (
                    <option key={option.id} value={option.id}>
                      {" ".repeat(option.path.length - 1)}
                      {option.title}
                    </option>
                  ))
                : parentId && (
                    <option value={parentId}>
                      {tree.status === "loading" ? "Betöltés…" : "Jelenlegi hely"}
                    </option>
                  )}
            </select>
          </Field>
        )}

        {error && (
          <p
            role="alert"
            className="rounded-control border border-brand-red/20 bg-brand-red/5 px-3.5 py-2.5 text-sm text-brand-red-strong"
          >
            {error}
          </p>
        )}

        <div className="flex justify-end gap-3 border-t border-line pt-5">
          <Button
            type="button"
            variant="secondary"
            onClick={props.onClose}
            disabled={submitting}
          >
            Mégse
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Mentés…" : editing ? "Mentés" : "Létrehozás"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
