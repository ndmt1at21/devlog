"use client";

import { useEffect, useRef } from "react";
import { track } from "@/lib/analytics";

const THRESHOLDS = [25, 50, 75, 100] as const;

/** Fires `scroll_depth` once per 25/50/75/100% threshold on an article page. */
export function ScrollDepthTracker({ slug }: { slug: string }) {
  const fired = useRef<Set<number>>(new Set());

  useEffect(() => {
    fired.current = new Set();
    const onScroll = () => {
      const h = document.documentElement;
      const max = h.scrollHeight - h.clientHeight;
      const pct = max > 0 ? (h.scrollTop / max) * 100 : 100;
      for (const t of THRESHOLDS) {
        if (pct >= t && !fired.current.has(t)) {
          fired.current.add(t);
          track("scroll_depth", { percent: t, slug });
        }
      }
    };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, [slug]);

  return null;
}
