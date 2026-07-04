"use client";

// A single Google AdSense display unit. Renders an <ins class="adsbygoogle"> and
// pushes it to the queue once on mount — adsbygoogle.js fills the next unrendered
// <ins> on each push. The pushed ref guards against React StrictMode's dev
// double-invoke (a second push on the same <ins> throws "already have ads").

import { useEffect, useRef } from "react";
import { ADSENSE_CLIENT, adSlotId, slotEnabled } from "@/lib/ads";

export function AdSenseSlot({ slot }: { slot: string }) {
  const slotId = adSlotId(slot);
  const pushed = useRef(false);

  useEffect(() => {
    if (!slotEnabled(slot) || pushed.current) return;
    pushed.current = true;
    try {
      (window.adsbygoogle = window.adsbygoogle || []).push({});
    } catch {
      // Script not ready yet or duplicate push — safe to ignore.
    }
  }, [slot]);

  if (!slotEnabled(slot)) return null;

  return (
    <ins
      className="adsbygoogle min-h-[90px] w-full"
      style={{ display: "block" }}
      data-ad-client={ADSENSE_CLIENT}
      data-ad-slot={slotId}
      data-ad-format="auto"
      data-full-width-responsive="true"
    />
  );
}
