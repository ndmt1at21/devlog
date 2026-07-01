// Typed, SSR-safe GA4 event wrapper. Every custom event goes through track() so
// instrumentation stays consistent; it no-ops on the server and when gtag isn't
// loaded (e.g. NEXT_PUBLIC_GA_ID unset).

type Gtag = (
  command: "event" | "config" | "js" | "set",
  target: string | Date,
  params?: Record<string, unknown>,
) => void;

declare global {
  interface Window {
    gtag?: Gtag;
    dataLayer?: unknown[];
  }
}

// The event taxonomy (see PLAN "Analytics & event tracking"). Union keeps call
// sites honest without forcing an exhaustive param schema per event.
export type AnalyticsEvent =
  | "view_article"
  | "select_article"
  | "search"
  | "select_category"
  | "select_tag"
  | "sign_up"
  | "login"
  | "logout"
  | "select_pro_plan"
  | "subscribe_pro"
  | "coffee_open"
  | "coffee_donate"
  | "copy_code"
  | "toggle_theme"
  | "series_nav"
  | "share_article"
  | "paywall_view"
  | "paywall_upgrade_click"
  | "ad_impression"
  | "ad_click"
  | "scroll_depth";

export function track(
  event: AnalyticsEvent,
  params: Record<string, unknown> = {},
): void {
  if (typeof window === "undefined" || typeof window.gtag !== "function") return;
  window.gtag("event", event, params);
}

export const GA_ID = process.env.NEXT_PUBLIC_GA_ID;
