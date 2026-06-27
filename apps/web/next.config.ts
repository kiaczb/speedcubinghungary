import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emit a self-contained production server at .next/standalone so the prod
  // Docker image can run without node_modules. No effect on `next dev`.
  output: "standalone",
  // Turbopack (default in Next 16) honors this for file watching. Polling is
  // needed because native FS events don't propagate over Docker bind mounts
  // in the dev container. Replaces the old webpack `watchOptions.poll`.
  watchOptions: {
    pollIntervalMs: 1000,
  },
};

export default nextConfig;
