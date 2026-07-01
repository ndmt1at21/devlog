"use client";

// Loads the Google Publisher Tag library and initializes the pubads service
// once. Renders nothing (and loads nothing) when GAM isn't configured.

import Script from "next/script";
import { gamEnabled, type GoogleTag } from "@/lib/ads";

export function GoogleAdManager() {
  if (!gamEnabled) return null;

  return (
    <Script
      id="gpt-js"
      src="https://securepubads.g.doubleclick.net/tag/js/gpt.js"
      strategy="afterInteractive"
      onReady={() => {
        const gt = (window.googletag =
          window.googletag || ({ cmd: [] } as unknown as GoogleTag));
        gt.cmd.push(() => {
          // No enableSingleRequest(): in-content slots are defined lazily as the
          // reader scrolls, so each fetches individually rather than up-front.
          gt.pubads().collapseEmptyDivs();
          gt.enableServices();
        });
      }}
    />
  );
}
