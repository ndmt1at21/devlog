// Centralized SEO constants and helpers. Used by metadata, JSON-LD, the sitemap
// and robots so the public origin and org identity are defined in exactly one
// place. Keep this free of `server-only` — the sitemap/robots route modules and
// (client) JSON-LD components all import from here.

/** Public site origin (no trailing slash), used for canonical/OG/sitemap URLs. */
export const SITE_URL = (
  process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"
).replace(/\/$/, "");

export const SITE_NAME = "devnote";

/** Resolve a site-relative path to an absolute URL. */
export function absoluteUrl(path = "/"): string {
  return `${SITE_URL}${path.startsWith("/") ? path : `/${path}`}`;
}

/** Publisher/author identity reused across Article & Organization JSON-LD. */
export const organization = {
  "@type": "Organization",
  name: SITE_NAME,
  url: SITE_URL,
  logo: {
    "@type": "ImageObject",
    url: absoluteUrl("/opengraph-image"),
  },
} as const;
