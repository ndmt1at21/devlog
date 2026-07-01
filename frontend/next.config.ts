import type { NextConfig } from "next";
import { initOpenNextCloudflareForDev } from "@opennextjs/cloudflare";

// Internal backend origin used for same-origin proxying. The browser only ever
// talks to Next at :3000, so httpOnly session cookies stay first-party. During
// SSR the API client targets BACKEND_INTERNAL_URL directly (see src/lib/api.ts).
// NOTE (Cloudflare): this rewrite destination is baked in at BUILD time, so
// BACKEND_INTERNAL_URL must be set in the build environment, not just as a
// runtime wrangler var.
const BACKEND = process.env.BACKEND_INTERNAL_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${BACKEND}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;

// Enables getCloudflareContext()/wrangler bindings during `next dev`. It's a
// no-op for `next build` and the OpenNext build.
initOpenNextCloudflareForDev();
