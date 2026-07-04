// OAuth passthrough for the Google login flow (/login and /callback). The
// generic /api/:path* rewrite (next.config.ts) is served by fetch() inside the
// Cloudflare Worker, which follows the backend's 302 to accounts.google.com
// server-side — the browser gets Google's HTML under our domain and the
// state/session Set-Cookie on the redirect response is dropped. These two
// endpoints are top-level browser navigations whose redirects and cookies must
// reach the browser untouched, so they are proxied here with redirect:
// "manual". App routes match before afterFiles rewrites, so all other /api
// traffic keeps using the rewrite.
import { NextRequest } from "next/server";

const actions = new Set(["login", "callback"]);

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ action: string }> },
) {
  const { action } = await params;
  if (!actions.has(action)) {
    return new Response("Not found", { status: 404 });
  }
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
