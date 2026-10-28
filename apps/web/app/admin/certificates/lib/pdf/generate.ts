import type { CertificateCompetition, CertificateMode, Recipient } from "../../types";
import type { CertificateDocument } from "../template/model";
import { loadDocumentAssets, loadDocumentFonts } from "../template/loaders";
import { certificateContexts } from "./content";
import { renderCertificates } from "./render";
import type { RenderedCertificates } from "./render";

export async function generateCertificates(
  document: CertificateDocument,
  competition: CertificateCompetition,
  mode: CertificateMode,
  recipients: Recipient[],
): Promise<RenderedCertificates> {
  const [fonts, assets] = await Promise.all([loadDocumentFonts(document), loadDocumentAssets(document)]);
  return renderCertificates(document, certificateContexts(competition, mode, recipients), fonts, assets, {
    title: `Oklevelek – ${competition.name}`,
    author: "Speedcubing Hungary",
  });
}
