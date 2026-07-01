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
} from "./types";

const INTERNAL = process.env.BACKEND_INTERNAL_URL ?? "http://localhost:8080";
const API_V1 = "/api/v1";

interface Envelope<T> {
  code: number;
  message: string;
  traceId: string;
  data: T;
}

async function serverGet<T>(path: string): Promise<T | null> {
  const cookieHeader = (await cookies()).toString();
  try {
    const res = await fetch(`${INTERNAL}${API_V1}${path}`, {
      headers: cookieHeader ? { cookie: cookieHeader } : undefined,
      cache: "no-store",
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
  return (await serverGet<ArticleSummary[]>(`/articles${suffix}`)) ?? [];
}

export async function fetchFeatured(): Promise<ArticleSummary | null> {
  return serverGet<ArticleSummary>("/articles/featured");
}

export async function fetchCategories(): Promise<string[]> {
  return (await serverGet<string[]>("/categories")) ?? ["Tất cả"];
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
