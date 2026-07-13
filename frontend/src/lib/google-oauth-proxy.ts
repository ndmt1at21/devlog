// Passthrough proxy for the Google OAuth endpoints (login/callback). The
// generic /api/:path* rewrite (next.config.ts) is served by fetch() inside the
// Cloudflare Worker, which follows the backend's 302 to accounts.google.com
// server-side — the browser gets Google's HTML under our domain and the
// state/session Set-Cookie on the redirect response is dropped. These two
// endpoints are top-level browser navigations whose redirects and cookies must
// reach the browser untouched, so they go through this manual-redirect proxy
// instead.
//
// The route files calling this MUST stay static (literal login/ and callback/
// segments, no [param]): afterFiles rewrites are matched BEFORE dynamic
// routes, so a dynamic segment here silently loses to the /api/:path* rewrite
// and the Worker follows the redirect again.
import "server-only";
import type { NextRequest } from "next/server";

export async function proxyGoogleAuth(
  req: NextRequest,
  action: "login" | "callback",
): Promise<Response> {
  const backend = process.env.BACKEND_INTERNAL_URL || "http://localhost:8080";
  const cookie = req.headers.get("cookie");
  const res = await fetch(
    `${backend}/api/v1/auth/google/${action}${req.nextUrl.search}`,
    {
      headers: cookie ? { cookie } : undefined,
      redirect: "manual",
      cache: "no-store",
    },
  );
  const headers = new Headers(res.headers);
  // The runtime already decompressed the body; the original envelope headers
  // would no longer match it.
  headers.delete("content-encoding");
  headers.delete("content-length");
  headers.delete("transfer-encoding");
  return new Response(res.body, { status: res.status, headers });
}
