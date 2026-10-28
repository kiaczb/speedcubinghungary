import { PDFDocument } from "pdf-lib";
import type { PDFFont } from "pdf-lib";
import fontkit from "@pdf-lib/fontkit";
import { loadAsset, loadFontFile, registerFontFaces } from "./loaders";
import { measurerFor } from "./metrics";
import type { Measurer } from "./textLayout";

let metricsDocument: Promise<PDFDocument> | null = null;
const embedded: Record<string, PDFFont> = {};
const pending = new Map<string, Promise<void>>();
const objectUrls = new Map<string, Promise<string>>();

function embed(file: string): Promise<void> {
  let task = pending.get(file);
  if (!task) {
    metricsDocument ??= PDFDocument.create().then((doc) => {
      doc.registerFontkit(fontkit);
      return doc;
    });
    const documentPromise = metricsDocument;
    task = (async () => {
      const doc = await documentPromise;
      embedded[file] = await doc.embedFont(await loadFontFile(file));
    })().catch((error) => {
      pending.delete(file);
      throw error;
    });
    pending.set(file, task);
  }
  return task;
}

export function hasFonts(files: string[]): boolean {
  return files.every((file) => file in embedded);
}

export async function editorMeasurer(files: string[]): Promise<Measurer> {
  await Promise.all([...files.map(embed), registerFontFaces(files)]);
  return measurerFor({ ...embedded });
}

export function assetObjectUrl(id: string): Promise<string> {
  let url = objectUrls.get(id);
  if (!url) {
    url = loadAsset(id).then(({ bytes, mime }) => URL.createObjectURL(new Blob([bytes], { type: mime })));
    url.catch(() => objectUrls.delete(id));
    objectUrls.set(id, url);
  }
  return url;
}
