// Client-side API helpers. In the browser these hit same-origin /api/v1/* paths,
// which Next rewrites to the Go backend (keeping the session cookie first-party).
// Every response is the uniform envelope { code, message, traceId, data }; this
// unwraps `data` on success (code 0) and throws a translated ApiError otherwise.
import type {
  ArticleDetail,
  ArticleSummary,
  Comment,
  MeResponse,
  NewArticleInput,
  Plan,
  ReactionStatus,
  SubscriptionState,
  UploadTicket,
} from "./types";
import { translateError } from "./errorCodes";

/** API version prefix. The Next rewrite forwards /api/* to the Go backend. */
export const API_BASE = "/api/v1";

interface Envelope<T> {
  code: number;
  message: string;
  traceId: string;
  data: T;
}

export class ApiError extends Error {
  status: number;
  /** Stable backend error code (see errorCodes.ts); 0 means success. */
  code: number;
  /** Server trace id for correlating with backend logs. */
  traceId?: string;
  constructor(message: string, status: number, code: number, traceId?: string) {
    super(message);
    this.status = status;
    this.code = code;
    this.traceId = traceId;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });
  const text = await res.text();
  let env: Envelope<T> | null = null;
  try {
    env = text ? (JSON.parse(text) as Envelope<T>) : null;
  } catch {
    throw new ApiError(translateError(undefined), res.status, 1006);
  }
  if (!env || env.code !== 0) {
    throw new ApiError(
      translateError(env?.code, env?.message),
      res.status,
      env?.code ?? 1006,
      env?.traceId,
    );
  }
  return env.data;
}

export const api = {
  // --- articles ---
  createArticle: (body: NewArticleInput) =>
    request<ArticleDetail>("/articles", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  // Edit an existing article (author-only; the backend re-checks ownership).
  updateArticle: (slug: string, body: NewArticleInput) =>
    request<ArticleDetail>(`/articles/${encodeURIComponent(slug)}`, {
      method: "PUT",
      body: JSON.stringify(body),
    }),

  // --- image uploads (presigned direct-to-bucket PUT) ---
  createUpload: (body: { type: string; size: number }) =>
    request<UploadTicket>("/uploads", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  // --- search ---
  searchArticles: (q: string) =>
    request<ArticleSummary[]>(`/articles?q=${encodeURIComponent(q)}`),

  // --- comments ---
  listComments: (slug: string) =>
    request<Comment[]>(`/articles/${encodeURIComponent(slug)}/comments`),
  createComment: (slug: string, body: { name: string; text: string }) =>
    request<Comment>(`/articles/${encodeURIComponent(slug)}/comments`, {
      method: "POST",
      body: JSON.stringify(body),
    }),

  // --- reactions (like / bookmark) ---
  reactions: (slug: string) =>
    request<ReactionStatus>(`/articles/${encodeURIComponent(slug)}/reactions`),
  setReaction: (slug: string, kind: "like" | "bookmark", on: boolean) =>
    request<ReactionStatus>(
      `/articles/${encodeURIComponent(slug)}/reactions/${kind}`,
      { method: on ? "PUT" : "DELETE" },
    ),

  // --- auth ---
  me: () => request<MeResponse>("/auth/me"),
  login: (body: { email: string; password: string }) =>
    request<{ authenticated: boolean }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  register: (body: { name: string; email: string; password: string }) =>
    request<{ status: string; message: string }>("/auth/register", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  forgotPassword: (body: { email: string }) =>
    request<{ ok: boolean }>("/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  logout: () => request<{ ok: boolean }>("/auth/logout", { method: "POST" }),

  // --- pro / subscription ---
  plans: () => request<Plan[]>("/pro/plans"),
  subscription: () => request<SubscriptionState>("/me/subscription"),
  subscribe: (plan: "month" | "year") =>
    request<SubscriptionState>("/me/subscription", {
      method: "POST",
      body: JSON.stringify({ plan }),
    }),

  // --- coffee ---
  coffeeCheckout: (body: {
    amount: number;
    method: "card" | "momo";
    name?: string;
    email?: string;
  }) =>
    request<{
      demo?: boolean;
      orderId?: string;
      redirectUrl?: string;
      qrCodeUrl?: string;
      deeplink?: string;
      payUrl?: string;
    }>("/coffee/checkout", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  coffeeStatus: (id: string) =>
    request<{ status: string; amount: number; method: string }>(
      `/coffee/${encodeURIComponent(id)}/status`,
    ),
};
