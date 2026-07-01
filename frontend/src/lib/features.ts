// Env-driven feature flags + donate config. Like ads.ts, every value here reads
// a NEXT_PUBLIC_* var, so it's inlined into the client bundle at BUILD time —
// these must be set in the build environment, not just as runtime wrangler vars.

// --- PRO subscription (the /pro page + every "Upgrade to Pro" entry point) ---
// On by default; set NEXT_PUBLIC_PRO_ENABLED=false to hide it everywhere (the
// /pro route then 404s, and upsell CTAs disappear).
export const proEnabled = process.env.NEXT_PUBLIC_PRO_ENABLED !== "false";

// --- Donate (PayPal + MoMo) ---
// Links are configured entirely via env; the modal makes no backend calls.
export const PAYPAL_DONATE_URL = process.env.NEXT_PUBLIC_PAYPAL_DONATE_URL ?? "";
export const MOMO_QR_URL = process.env.NEXT_PUBLIC_MOMO_QR_URL ?? "";
export const MOMO_DONATE_URL = process.env.NEXT_PUBLIC_MOMO_DONATE_URL ?? "";

// A channel shows only when it's configured: PayPal needs a link, MoMo needs a
// QR image (the optional deep link just adds an "Open MoMo" button on top).
export const hasPaypal = PAYPAL_DONATE_URL !== "";
export const hasMomo = MOMO_QR_URL !== "";

// Donate is offered when not force-disabled AND at least one channel exists — so
// with nothing configured the button auto-hides rather than opening an empty modal.
export const donateEnabled =
  process.env.NEXT_PUBLIC_DONATE_ENABLED !== "false" && (hasPaypal || hasMomo);
