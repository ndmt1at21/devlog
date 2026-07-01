import type { MetadataRoute } from "next";
import type { ArticleSummary } from "@/lib/types";
import { SITE_URL } from "@/lib/seo";

const INTERNAL = process.env.BACKEND_INTERNAL_URL ?? "http://localhost:8080";

// The backend wraps every response in { code, message, traceId, data }.
interface Envelope<T> {
  code: number;
  data: T;
}

export const dynamic = "force-dynamic";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  let articles: ArticleSummary[] = [];
  try {
    // NOTE: the API is versioned (/api/v1) and enveloped — earlier this hit the
    // wrong path and never unwrapped `data`, so no article URLs were emitted.
    const res = await fetch(`${INTERNAL}/api/v1/articles`, {
      cache: "no-store",
    });
    if (res.ok) {
      const env = (await res.json()) as Envelope<ArticleSummary[]>;
      if (env.code === 0 && Array.isArray(env.data)) articles = env.data;
    }
  } catch {
    /* backend unavailable — emit static routes only */
  }

  const now = new Date();
  const staticRoutes: MetadataRoute.Sitemap = [
    { url: SITE_URL, lastModified: now, changeFrequency: "daily", priority: 1 },
    {
      url: `${SITE_URL}/pro`,
      lastModified: now,
      changeFrequency: "monthly",
      priority: 0.5,
    },
  ];

  const articleRoutes: MetadataRoute.Sitemap = articles.map((a) => ({
    url: `${SITE_URL}/articles/${a.slug}`,
    lastModified: a.publishedAt ? new Date(a.publishedAt) : now,
    changeFrequency: "weekly",
    priority: 0.8,
  }));

  return [...staticRoutes, ...articleRoutes];
}
