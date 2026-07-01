import { defineCloudflareConfig } from "@opennextjs/cloudflare";

// This app is fully dynamic (per-request cookies, `cache: "no-store"`), so no
// incremental/ISR cache is needed. If you later add ISR or `revalidate`, wire
// an incrementalCache here (e.g. R2/KV) — see the OpenNext Cloudflare docs.
export default defineCloudflareConfig({});
