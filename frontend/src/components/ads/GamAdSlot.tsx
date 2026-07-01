"use client";

// A single Google Ad Manager ad unit. Defines + displays the slot on mount and
// destroys it on unmount (so it's cleaned up on client-side route changes).
// The command queue means this works whether or not gpt.js has finished loading.

import { useEffect, useId, useRef } from "react";
import {
  AD_SIZES,
  adUnitPath,
  gamEnabled,
  type GoogleTag,
  type GptSlot,
} from "@/lib/ads";

export function GamAdSlot({ slot }: { slot: string }) {
  // useId is stable across SSR/CSR; strip ":" so it's a valid DOM id.
  const divId = `gam-${slot}-${useId().replace(/:/g, "")}`;
  const slotRef = useRef<GptSlot | null>(null);

  useEffect(() => {
    if (!gamEnabled) return;
    const gt = (window.googletag =
      window.googletag || ({ cmd: [] } as unknown as GoogleTag));

    gt.cmd.push(() => {
      const defined = gt.defineSlot(adUnitPath(slot), AD_SIZES, divId);
      if (!defined) return;
      // Responsive: leaderboard on >=768px viewports, rectangle below.
      const mapping = gt
        .sizeMapping()
        .addSize([768, 0], [[728, 90]])
        .addSize([0, 0], [[300, 250]])
        .build();
      if (mapping) defined.defineSizeMapping(mapping);
      defined.setCollapseEmptyDiv(true);
      defined.addService(gt.pubads());
      slotRef.current = defined;
      gt.display(divId);
    });

    return () => {
      const s = slotRef.current;
      if (s && window.googletag) {
        window.googletag.cmd.push(() => window.googletag!.destroySlots([s]));
        slotRef.current = null;
      }
    };
  }, [slot, divId]);

  return <div id={divId} className="min-h-[90px] w-full" />;
}
