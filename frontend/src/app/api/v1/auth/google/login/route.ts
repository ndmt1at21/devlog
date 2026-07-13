import type { NextRequest } from "next/server";
import { proxyGoogleAuth } from "@/lib/google-oauth-proxy";

export function GET(req: NextRequest): Promise<Response> {
  return proxyGoogleAuth(req, "login");
}
