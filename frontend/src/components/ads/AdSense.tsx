"use client";

// Loads the Google AdSense library once, tagged with the publisher id. Renders
// nothing (and loads nothing) when AdSense isn't configured. Individual ad units
// are placed by <AdSenseSlot>; the script just needs to be present on the page.

import Script from "next/script";
import { ADSENSE_CLIENT, adsenseEnabled } from "@/lib/ads";

export function AdSense() {
  if (!adsenseEnabled) return null;

  return (
    <Script
      id="adsbygoogle-js"
      src={`https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${ADSENSE_CLIENT}`}
      strategy="afterInteractive"
      crossOrigin="anonymous"
    />
  );
}
