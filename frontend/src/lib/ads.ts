// Google AdSense configuration. Everything is gated on the publisher (client)
// id: with it unset, `adsenseEnabled` is false and the UI keeps the placeholder
// slot — so nothing loads and no requests are made until you provide real IDs.

export const ADSENSE_CLIENT = process.env.NEXT_PUBLIC_ADSENSE_CLIENT ?? "";
export const adsenseEnabled = ADSENSE_CLIENT !== "";

// Logical slot name → AdSense ad-unit slot id. Create a display ad unit per slot
// in your AdSense account; each one gives you the numeric `data-ad-slot` value.
// Unknown/unconfigured slots return "" and render nothing (placeholder stays).
const AD_SLOT_IDS: Record<string, string> = {
  "in-content": process.env.NEXT_PUBLIC_ADSENSE_SLOT ?? "",
};

export function adSlotId(slot: string): string {
  return AD_SLOT_IDS[slot] ?? "";
}

// True when both the publisher id and this slot's ad-unit id are configured.
export function slotEnabled(slot: string): boolean {
  return adsenseEnabled && adSlotId(slot) !== "";
}

// ---- Minimal adsbygoogle typing (only what we use) ----

declare global {
  interface Window {
    // Each `push({})` renders the next unrendered <ins class="adsbygoogle">.
    adsbygoogle?: Array<Record<string, unknown>>;
  }
}
