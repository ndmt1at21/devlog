// Google Ad Manager (GPT) configuration. Everything is gated on the network
// code: with it unset, `gamEnabled` is false and the UI keeps the placeholder
// slot — so nothing loads and no requests are made until you provide real IDs.

export const GAM_NETWORK_CODE = process.env.NEXT_PUBLIC_GAM_NETWORK_CODE ?? "";
export const gamEnabled = GAM_NETWORK_CODE !== "";

// Logical slot name → GAM ad-unit name (configure these units in your GAM
// network). The full ad-unit path is `/{networkCode}/{unitName}`.
const AD_UNIT_NAMES: Record<string, string> = {
  "in-content": process.env.NEXT_PUBLIC_GAM_AD_UNIT ?? "jamti_in_content",
};

export function adUnitPath(slot: string): string {
  const name = AD_UNIT_NAMES[slot] ?? `jamti_${slot.replace(/-/g, "_")}`;
  return `/${GAM_NETWORK_CODE}/${name}`;
}

// Design shows a 728×90 leaderboard; 300×250 is the mobile fallback (applied via
// the size mapping in GamAdSlot). The union of both is the slot's size list.
export const AD_SIZES: Array<[number, number]> = [
  [728, 90],
  [300, 250],
];

// ---- Minimal Google Publisher Tag typings (only what we use) ----

export interface GptSlot {
  addService(service: GptPubAdsService): GptSlot;
  defineSizeMapping(mapping: unknown[]): GptSlot;
  setCollapseEmptyDiv(collapse: boolean): GptSlot;
}

interface GptPubAdsService {
  enableSingleRequest(): void;
  collapseEmptyDivs(): void;
  refresh(slots?: GptSlot[]): void;
}

interface GptSizeMappingBuilder {
  addSize(
    viewport: [number, number],
    sizes: Array<[number, number]>,
  ): GptSizeMappingBuilder;
  build(): unknown[] | null;
}

export interface GoogleTag {
  cmd: Array<() => void>;
  defineSlot(
    path: string,
    sizes: Array<[number, number]>,
    divId: string,
  ): GptSlot | null;
  pubads(): GptPubAdsService;
  enableServices(): void;
  display(divId: string): void;
  destroySlots(slots?: GptSlot[]): boolean;
  sizeMapping(): GptSizeMappingBuilder;
}

declare global {
  interface Window {
    googletag?: GoogleTag;
  }
}
