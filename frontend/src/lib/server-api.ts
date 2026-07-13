// Server-only API client. Runs in Server Components / route handlers: it targets
// the Go backend directly (BACKEND_INTERNAL_URL) and forwards the incoming
// session cookie so the backend can apply the paywall / premium gating per user.
// Responses use the uniform envelope { code, message, traceId, data }; this
// returns `data` on success (code 0) and null on any error.
import "server-only";
import { cookies } from "next/headers";
import type {
  ArticleDetail,
  ArticleSummary,
  Comment,
  MeResponse,
  ReactionStatus,
} from "./types";

const INTERNAL = process.env.BACKEND_INTERNAL_URL ?? "http://localhost:8080";
const API_V1 = "/api/v1";

/** Public content is served from the Next data cache (ISR) for this long. */
const CONTENT_REVALIDATE_S = 60;

interface Envelope<T> {
  code: number;
  message: string;
  traceId: string;
  data: T;
}

/**
 * Per-user fetch: forwards the session cookie and is never cached. Required for
 * anything whose response depends on the requester — auth, reactions,
 * bookmarks, and the article detail (its body is truncated by the paywall
 * according to the user's premium status).
 *
 * These run during a Server Component render, which cannot write cookies back
 * to the browser. The `X-Session-Read` header tells the backend to treat the
 * request as read-only: it resolves the caller from the sealed session cookie
 * but never rotates the refresh token, so a mid-flight token rotation here can't
 * revoke the cookie the browser still holds. The browser's own /auth/me
 * revalidation (a first-party request) is what actually refreshes the session.
 */
async function serverGet<T>(path: string): Promise<T | null> {
  const cookieHeader = (await cookies()).toString();
  try {
    const headers: Record<string, string> = { "X-Session-Read": "1" };
    if (cookieHeader) headers.cookie = cookieHeader;
    const res = await fetch(`${INTERNAL}${API_V1}${path}`, {
      headers,
      cache: "no-store",
    });
    const env = (await res.json()) as Envelope<T>;
    if (!res.ok || env.code !== 0) return null;
    return env.data;
  } catch {
    return null;
  }
}

/**
 * Shared-content fetch (ISR): identical for every visitor, so no cookie is
 * sent (a per-user header would poison the shared cache) and the result is
 * reused across requests for CONTENT_REVALIDATE_S seconds.
 */
async function cachedGet<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${INTERNAL}${API_V1}${path}`, {
      next: { revalidate: CONTENT_REVALIDATE_S },
    });
    const env = (await res.json()) as Envelope<T>;
    if (!res.ok || env.code !== 0) return null;
    return env.data;
  } catch {
    return null;
  }
}

export async function fetchArticles(params?: {
  category?: string;
  q?: string;
}): Promise<ArticleSummary[]> {
  const qs = new URLSearchParams();
  if (params?.category) qs.set("category", params.category);
  if (params?.q) qs.set("q", params.q);
  const suffix = qs.toString() ? `?${qs.toString()}` : "";
  return (await cachedGet<ArticleSummary[]>(`/articles${suffix}`)) ?? [];
}

export async function fetchFeatured(): Promise<ArticleSummary[]> {
  return (await cachedGet<ArticleSummary[]>("/articles/featured")) ?? [];
}

export async function fetchCategories(): Promise<string[]> {
  return (await cachedGet<string[]>("/categories")) ?? ["Tất cả"];
}

export async function fetchArticle(slug: string): Promise<ArticleDetail | null> {
  return serverGet<ArticleDetail>(`/articles/${encodeURIComponent(slug)}`);
}

export async function fetchComments(slug: string): Promise<Comment[]> {
  return (
    (await serverGet<Comment[]>(
      `/articles/${encodeURIComponent(slug)}/comments`,
    )) ?? []
  );
}

export async function fetchMe(): Promise<MeResponse> {
  return (await serverGet<MeResponse>("/auth/me")) ?? { authenticated: false };
}

export async function fetchReactions(slug: string): Promise<ReactionStatus> {
  return (
    (await serverGet<ReactionStatus>(
      `/articles/${encodeURIComponent(slug)}/reactions`,
    )) ?? { likes: 0, liked: false, bookmarked: false }
  );
}

/** The signed-in user's saved articles; null when not authenticated. */
export async function fetchBookmarks(): Promise<ArticleSummary[] | null> {
  return serverGet<ArticleSummary[]>("/me/bookmarks");
}
