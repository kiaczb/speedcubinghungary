import { clientFetch } from "@/app/lib/api";
import type { CertificateDocument } from "./model";
import { FONT_BASE_URL, cssFontFamily, documentFontFiles } from "./fonts";
import type { FontBytes } from "./metrics";

const fontCache = new Map<string, Promise<ArrayBuffer>>();
const assetCache = new Map<string, Promise<{ bytes: ArrayBuffer; mime: string }>>();
const registeredFaces = new Set<string>();

function remember<T>(cache: Map<string, Promise<T>>, key: string, load: () => Promise<T>): Promise<T> {
  let entry = cache.get(key);
  if (!entry) {
    entry = load().catch((error) => {
      cache.delete(key);
      throw error;
    });
    cache.set(key, entry);
  }
  return entry;
}

export function loadFontFile(file: string): Promise<ArrayBuffer> {
  return remember(fontCache, file, async () => {
    const response = await fetch(`${FONT_BASE_URL}${file}`);
    if (!response.ok) throw new Error(`A betűkészlet nem tölthető le: ${file} (${response.status})`);
    return response.arrayBuffer();
  });
}

export async function loadFonts(files: string[]): Promise<FontBytes> {
  const buffers = await Promise.all(files.map(loadFontFile));
  return Object.fromEntries(files.map((file, i) => [file, buffers[i]]));
}

export function loadDocumentFonts(document: CertificateDocument): Promise<FontBytes> {
  return loadFonts(documentFontFiles(document));
}

export function assetUrl(id: string): string {
  return `/certificates/assets/${encodeURIComponent(id)}`;
}

export function loadAsset(id: string): Promise<{ bytes: ArrayBuffer; mime: string }> {
  return remember(assetCache, id, async () => {
    const response = await clientFetch(assetUrl(id));
    return {
      bytes: await response.arrayBuffer(),
      mime: response.headers.get("Content-Type") ?? "image/png",
    };
  });
}

export async function loadDocumentAssets(document: CertificateDocument) {
  const ids = [
    ...new Set(
      document.elements.flatMap((e) => (e.type === "image" && e.assetId && !e.hidden ? [e.assetId] : [])),
    ),
  ];
  const loaded = await Promise.allSettled(ids.map(loadAsset));
  return Object.fromEntries(
    ids.flatMap((id, i) => {
      const result = loaded[i];
      return result.status === "fulfilled" ? [[id, result.value]] : [];
    }),
  );
}

export async function registerFontFaces(files: string[]): Promise<void> {
  await Promise.all(
    files
      .filter((file) => !registeredFaces.has(file))
      .map(async (file) => {
        const face = new FontFace(cssFontFamily(file), (await loadFontFile(file)).slice(0));
        await face.load();
        document.fonts.add(face);
        registeredFaces.add(file);
      }),
  );
}
