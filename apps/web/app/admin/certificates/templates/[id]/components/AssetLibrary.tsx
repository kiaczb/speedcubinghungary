"use client";

import { useRef, useState } from "react";
import { FaRegTrashAlt, FaUpload } from "react-icons/fa";
import Button from "@/app/components/ui/Button";
import type { CertificateAsset } from "../../../lib/template/model";
import { DRAG_TYPE } from "../lib/elements";

export default function AssetLibrary({
  assets,
  assetUrls,
  selectedId,
  onPick,
  onUpload,
  onDelete,
  draggable,
}: {
  assets: CertificateAsset[];
  assetUrls: Record<string, string>;
  selectedId?: string | null;
  onPick?: (asset: CertificateAsset) => void;
  onUpload: (file: File) => Promise<void>;
  onDelete?: (asset: CertificateAsset) => void;
  draggable?: boolean;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const upload = async (file: File) => {
    setUploading(true);
    setError(null);
    try {
      await onUpload(file);
    } catch (err) {
      setError(err instanceof Error ? err.message : "A feltöltés nem sikerült.");
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  };

  return (
    <div className="flex flex-col gap-2">
      {assets.length > 0 && (
        <ul className="grid grid-cols-3 gap-2">
          {assets.map((asset) => (
            <li key={asset.id} className="group relative">
              <button
                type="button"
                title={`${asset.name} · ${asset.width}×${asset.height} px`}
                draggable={draggable}
                onDragStart={(event) =>
                  event.dataTransfer.setData(
                    DRAG_TYPE,
                    JSON.stringify({ type: "image", assetId: asset.id, aspect: asset.width / asset.height }),
                  )
                }
                onClick={() => onPick?.(asset)}
                className={`flex aspect-square w-full items-center justify-center overflow-hidden rounded-md border bg-[repeating-conic-gradient(#f2f4f7_0_25%,#ffffff_0_50%)] bg-[length:12px_12px] p-1 transition-colors duration-200 ease-smooth hover:border-foreground ${
                  selectedId === asset.id ? "border-foreground ring-2 ring-foreground/20" : "border-line"
                }`}
              >
                {assetUrls[asset.id] ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={assetUrls[asset.id]} alt={asset.name} className="max-h-full max-w-full object-contain" />
                ) : (
                  <span className="text-[10px] text-subtle-foreground">…</span>
                )}
              </button>
              {onDelete && (
                <button
                  type="button"
                  aria-label={`${asset.name} törlése`}
                  onClick={() => onDelete(asset)}
                  className="absolute right-1 top-1 hidden h-6 w-6 items-center justify-center rounded-md bg-surface/90 text-muted-foreground shadow-card hover:text-brand-red group-hover:flex"
                >
                  <FaRegTrashAlt className="h-3 w-3" />
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg"
        className="hidden"
        onChange={(event) => {
          const file = event.target.files?.[0];
          if (file) upload(file);
        }}
      />
      <Button
        type="button"
        size="sm"
        variant="secondary"
        disabled={uploading}
        onClick={() => inputRef.current?.click()}
        icon={<FaUpload className="h-3 w-3" />}
      >
        {uploading ? "Feltöltés…" : "Kép feltöltése"}
      </Button>
      <p className="text-xs text-muted-foreground">PNG vagy JPEG, legfeljebb 10 MB. Nyomtatáshoz 300 dpi az ideális.</p>
      {error && <p className="text-xs text-brand-red">{error}</p>}
    </div>
  );
}
